#! /usr/bin/env bash

# Requires bash 4.3+ (declare -A, local -n).
if [ "${BASH_VERSINFO[0]}" -lt 4 ] ||
  { [ "${BASH_VERSINFO[0]}" -eq 4 ] && [ "${BASH_VERSINFO[1]}" -lt 3 ]; }; then
  echo "main.sh requires bash >= 4.3 (found $BASH_VERSION)" >&2
  exit 1
fi
cd -P "$(dirname "$0")" || exit 1

# ---- internal constants (not user configuration) ----
tmux_root=tasks          # tmux-mode root dir (relative to $HOME)
systemd_root=/srv/http   # systemd-mode root dir
sys=/etc/systemd/system  # systemd unit dir
sudoers_name=webdeploy   # shared per-user sudoers file suffix (do not change)

# ---- user configuration ----
[ -f deploy.conf ] || {
  echo "deploy.conf missing — restore it from git" >&2
  exit 1
}
source ./deploy.conf
use_rsync="${USE_RSYNC:-$use_rsync}"

# ---- derived / runtime state (do not edit) ----
tmux_dir="$tmux_root/$pname"       # tmux deploy dir
systemd_dir="$systemd_root/$pname" # systemd deploy dir
tmp_dir=""                         # packaging temp dir (created on demand)
# Absolute remote binary paths, resolved once per target (sudoers needs exact
# paths). Filled by resolve_remote_bins.
remote_systemctl=""
remote_install=""
remote_journalctl=""
remote_nginx=""
remote_caddy=""
remote_tee=""
remote_mkdir=""
srv_name=""                        # selected server (set by select_server)
srv_user=""
srv_host=""
ssh_target=""               # ssh target
SUDO=""                     # remote command prefix (root skips sudo)
build_mode="build"          # build | deploy (drives placeholder-env strictness)
allow_placeholder_env=false # set by deploy --allow-placeholder-env
current_cmd=""              # dispatched command (drives on_interrupt)

success() { echo -e "\033[0;32m[SUCCESS]\033[0m $1"; }
info() { echo -e "\033[1;36m[INFO]\033[0m $1"; }
warn() { echo -e "\033[1;33m[WARN]\033[0m $1"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1"; }

remote() {
  # shellcheck disable=SC2029
  ssh "$ssh_target" "$@"
}

remote_util() {
  local src="$1" target="$2"
  if [ -z "$target" ]; then
    error "remote_util: missing target directory (2nd arg required)."
    return 1
  fi
  if [ "$use_rsync" = true ]; then
    command -v rsync >/dev/null 2>&1 || {
      error "use_rsync=true but rsync is not installed locally"
      return 1
    }
    info "use rsync to sync files..."
    rsync -avP "$src" "$ssh_target:$target" || {
      error "rsync failed: $src -> $ssh_target:$target"
      return 1
    }
  else
    command -v scp >/dev/null 2>&1 || {
      error "use_rsync=false but scp is not installed locally"
      return 1
    }
    info "use scp to sync files..."
    scp "$src" "$ssh_target:$target" || {
      error "scp failed: $src -> $ssh_target:$target"
      return 1
    }
  fi
}

run_in_dir() {
  local dir="$1"
  shift
  if [ ! -d "$dir" ]; then
    error "Directory not found: $dir"
    return 1
  fi
  (cd "$dir" && "$@")
}

# Create the packaging temp dir on first use (usage/servers leave none behind).
ensure_tmp_dir() {
  [ -n "$tmp_dir" ] && return 0
  tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/${pname}.XXXXXX") || {
    error "cannot create temp dir"
    tmp_dir=""
    return 1
  }
}

# Select a target server from the servers[] registry (default: first entry).
select_server() {
  local name="${1:-${servers[0]%%|*}}" entry
  for entry in "${servers[@]}"; do
    IFS='|' read -r srv_name srv_user srv_host <<<"$entry"
    if [ "$srv_name" = "$name" ]; then
      ssh_target="$srv_user@$srv_host"
      SUDO=""
      [ "$srv_user" = "root" ] || SUDO="sudo" # root skips sudo, others use it
      info "Target server: $srv_name ($ssh_target)"
      return 0
    fi
  done
  local list=""
  for entry in "${servers[@]}"; do list+="${entry%%|*} "; done
  error "Unknown server '$name'. Available: $list"
  return 1
}

is_server() {
  local name="$1" entry
  for entry in "${servers[@]}"; do
    [ "${entry%%|*}" = "$name" ] && return 0
  done
  return 1
}

list() {
  local entry name path cmd out_dir srv_user srv_host
  info "Available servers (default: ${servers[0]%%|*}):"
  for entry in "${servers[@]}"; do
    IFS='|' read -r name srv_user srv_host <<<"$entry"
    echo "  $name  ($srv_user@$srv_host)"
  done
  info "Available apps:"
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    echo "  $name  (dir: $path, cmd: $cmd, out: $out_dir)"
  done
  info "Deploy mode: $deploy_mode (web server: $web_server, transfer: $use_rsync)"
  info "Domains:"
  local ep d shown=false
  while IFS= read -r ep; do
    d=$(domain_of "$ep")
    [ -n "$d" ] || continue
    echo "  $ep  ->  $d"
    shown=true
  done < <(endpoints)
  [ "$shown" = true ] || echo "  (none set in deploy.conf)"
}

dev() {
  current_cmd=dev
  command -v tmux >/dev/null 2>&1 || {
    error "tmux not installed"
    return 1
  }
  if [ "$2" = "restart" ]; then
    warn "Restarting tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
  fi
  if [ "$2" = "stop" ]; then
    warn "Stopping tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
    return 0
  fi
  if tmux has-session -t "$pname" 2>/dev/null; then
    warn "Session $pname already exists."
    return
  fi
  tmux new-session -d -s "$pname" -n "go" && tmux send-keys -t "$pname:0" 'gowatch -o bin/backend' C-m
  local i=1 entry name path
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path _ _ <<<"$entry"
    tmux new-window -t "$pname:$i" -n "$name" &&
      tmux send-keys -t "$pname:$i" "cd $(printf '%q' "$path") && pnpm dev" C-m
    i=$((i + 1))
  done
  info "Dev environment started in tmux session: $pname"
}

# Generate docs/swagger.json (+ docs/docs.go) from the swag annotations.
gen_swagger() {
  command -v swag >/dev/null 2>&1 || {
    error "swag not installed — run ./main.sh init"
    return 1
  }
  swag fmt >/dev/null || warn "swag fmt failed (continuing)"
  # --parseDependency walks the whole dependency graph, so swag also evaluates
  # consts from the Go standard library. Go 1.27 declares runtime/mprof.go's
  # mProfCycleWrap with a variable-length array that swag cannot fold, printing
  # one `failed to evaluate const` line on every run. The generated doc is
  # unaffected (verified: swagger.json is byte-identical with or without the
  # line), so drop only that known-harmless line and keep every other swag
  # diagnostic. swag's logger writes to stdout, hence the pipe; PIPESTATUS[0]
  # keeps swag's own exit status instead of grep's (grep exits 1 when it
  # filters out everything).
  swag init --parseDependency --parseInternal |
    grep -v 'failed to evaluate const mProfCycleWrap'
  local rc=${PIPESTATUS[0]}
  if [ "$rc" -ne 0 ]; then
    error "swag init failed"
    return 1
  fi
}

# Regenerate the shared typed SDK (web/packages/sdk/src/{usr,adm}).
# `alova gen -f` rewrites every artifact on each run and the facade/entry points
# are emitted by our own plugins, so a stale generated file cannot survive here.
gen_sdk() {
  run_in_dir web pnpm gen:api || {
    error "SDK generation failed"
    return 1
  }
  # Every app in apps[] must have produced its generated entry points.
  local entry name path cmd out_dir f missing=0
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    for f in "web/packages/sdk/src/$name/api.ts" "web/packages/sdk/src/$name/index.ts"; do
      if [ ! -f "$f" ]; then
        error "SDK generation produced no $f"
        missing=1
      fi
    done
  done
  return $missing
}

# Build backend + SDK (server-independent). Run once per deploy.
#
# docs/ is gitignored, and internal/srv/swg imports package docs, so a clean
# clone cannot compile until `swag init` has run. Swagger (and therefore the SDK)
# is generated first, then `go build` and `pnpm gen:api` run in parallel.
build_core() {
  info "Building backend & SDK"
  if command -v swag >/dev/null 2>&1; then
    gen_swagger || return 1
  elif [ -f docs/docs.go ]; then
    warn "swag not installed, reusing existing docs/ (may be stale)."
  else
    error "docs/docs.go missing and swag is not installed — run ./main.sh init."
    return 1
  fi

  mkdir -p bin
  local rc=0 p1
  go build -o bin/backend &
  p1=$!
  gen_sdk || rc=1
  wait "$p1" || rc=1
  return $rc
}

# Read `<VAR>=` from an app's .env.production (last occurrence, no quote support).
# usage: file_env <app dir> <VAR>
file_env() {
  local f="$1/.env.production" line
  [ -f "$f" ] || return 0
  line=$(grep -E "^${2}=" "$f" | tail -1) || true
  [ -n "$line" ] && printf '%s' "${line#*=}"
  return 0
}

# Print the value of VAR from a NAME=value array; fails when absent.
# usage: env_value <VAR> <array name>
env_value() {
  local var="$1" v
  local -n _a="$2"
  for v in "${_a[@]}"; do
    [ "${v%%=*}" = "$var" ] && {
      printf '%s' "${v#*=}"
      return 0
    }
  done
  return 1
}

# Collect an app's build-time env: explicit app_env entries for the selected
# server first, then the auto-derived NUXT_PUBLIC_* values (explicit wins).
# usage: app_envs <app> <array name>
app_envs() {
  local app="$1" k
  local -n _out="$2"
  _out=()
  for k in "${!app_env[@]}"; do
    [[ "$k" == "$srv_name:$app:"* ]] && _out+=("${k#"$srv_name:$app:"}=${app_env[$k]}")
  done
  auto_envs "$app" "$2"
}

# Guard against shipping a bundle that points at a placeholder origin.
#
# When domain_api / domain_usr / domain_adm are empty and no app_env override
# exists, the build falls back to .env.production, which ships localhost /
# example.com. `deploy` refuses such a bundle unless the operator opts out with
# --allow-placeholder-env; `build` only warns.
# usage: check_public_env <app> <build|deploy> <api_base> <site_url>
check_public_env() {
  local app="$1" mode="$2" base="$3" site="$4" bad=0
  [ "$base" = "same-origin" ] || case "$base" in
  "" | *localhost* | *127.0.0.1* | *example.com* | *example.org* | *example.net*) bad=1 ;;
  esac
  case "$site" in
  "" | *localhost* | *127.0.0.1* | *example.com* | *example.org* | *example.net*) bad=2 ;;
  esac
  [ $bad -eq 0 ] && return 0
  local msg
  if [ $bad -eq 1 ]; then
    msg="app '$app' would ship NUXT_PUBLIC_API_BASE='${base:-<unset>}'"
  else
    msg="app '$app' would ship NUXT_PUBLIC_SITE_URL='${site:-<unset>}'"
  fi
  if [ "$mode" = "deploy" ] && [ "$allow_placeholder_env" != true ]; then
    error "$msg"
    printf '  Fix it in one of these ways:\n' >&2
    printf '    1. set domain_api / domain_usr / domain_adm in deploy.conf\n' >&2
    printf '    2. add app_env["%s:%s:VARNAME"]=https://your-domain in deploy.conf\n' "$srv_name" "$app" >&2
    printf '    3. use NUXT_PUBLIC_API_BASE=same-origin when frontend and API share a host\n' >&2
    printf '  Or re-run with --allow-placeholder-env to deploy anyway.\n' >&2
    return 1
  fi
  warn "$msg — fine for local dev, not for production."
  return 0
}

# Build every app in parallel, exporting the selected server's env overrides.
# The marker written next to each build records what the bundle was built for,
# so --skip-build can refuse an artifact that no longer matches the target.
build_front() {
  info "Building frontends for $srv_name ..."
  local entry name path cmd out_dir
  local -a cmd_args=() pids=() names=() paths=() outs=() api_vals=() site_vals=()
  local rc=0 p i api site

  # phase 1: resolve + validate every app before starting any build
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ -d "$path" ] || {
      error "App directory not found: $path"
      return 1
    }
    local -a envs=()
    app_envs "$name" envs
    api=$(env_value NUXT_PUBLIC_API_BASE envs || file_env "$path" NUXT_PUBLIC_API_BASE)
    site=$(env_value NUXT_PUBLIC_SITE_URL envs || file_env "$path" NUXT_PUBLIC_SITE_URL)
    check_public_env "$name" "$build_mode" "$api" "$site" || return 1
    names+=("$name")
    paths+=("$path")
    outs+=("$out_dir")
    api_vals+=("$api")
    site_vals+=("$site")
  done

  # phase 2: build; `same-origin` is a main.sh keyword, not a URL
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    local -a envs=()
    app_envs "$name" envs
    for i in "${!envs[@]}"; do
      [ "${envs[$i]}" = "NUXT_PUBLIC_API_BASE=same-origin" ] && envs[i]="NUXT_PUBLIC_API_BASE="
    done
    read -ra cmd_args <<<"$cmd"
    info "  building $name ($path)"
    if [ ${#envs[@]} -gt 0 ]; then
      (cd "$path" && env "${envs[@]}" "${cmd_args[@]}") &
    else
      (cd "$path" && "${cmd_args[@]}") &
    fi
    pids+=($!)
  done
  for p in "${pids[@]}"; do wait "$p" || rc=1; done
  [ $rc -eq 0 ] || return $rc

  # phase 3: verify the output exists, then record what it was built for
  for i in "${!names[@]}"; do
    if [ ! -f "${paths[$i]}/${outs[$i]}/public/index.html" ]; then
      error "${names[$i]} build produced no ${paths[$i]}/${outs[$i]}/public/index.html"
      return 1
    fi
    printf 'server=%s\nNUXT_PUBLIC_API_BASE=%s\nNUXT_PUBLIC_SITE_URL=%s\n' \
      "$srv_name" "${api_vals[$i]}" "${site_vals[$i]}" >"${paths[$i]}/${outs[$i]}/.built-for" || return 1
  done
  return 0
}

# Backend + SDK, then frontends for the selected server.
#
# These run in sequence, not in parallel: `pnpm generate` imports the generated
# SDK, so building the frontends while gen:api rewrites app/lib/sdk races and can
# fail or bake in a half-written SDK.
build() {
  select_server "${2:-}" || return 1
  build_core || return 1
  build_front
}

deploy() {
  local skip_build=false force_service=false arg
  build_mode="deploy"
  for arg in "$@"; do
    case "$arg" in
    --skip-build) skip_build=true ;;
    --allow-placeholder-env) allow_placeholder_env=true ;;
    --force-service)
      if [ "$deploy_mode" != "systemctl" ]; then
        warn "--force-service only applies in deploy_mode=systemctl, ignored"
      else
        force_service=true
      fi
      ;;
    -*)
      error "Unknown deploy option: $arg"
      usage >&2
      return 1
      ;;
    *)
      # A stray positional used to be ignored, so `deploy prod` shipped to every
      # server in servers[] instead of the one the operator named.
      error "Unexpected deploy argument: $arg (deploy ships to every server in servers[]; use ./main.sh build [server])"
      return 1
      ;;
    esac
  done

  # backend + SDK are built only once, they are server-independent
  if [ "$skip_build" = false ]; then
    build_core || {
      error "Backend build failed, aborting deploy."
      return 1
    }
  fi

  # deploy to every server in servers[] sequentially (frontends rebuilt per-server env)
  local rc=0 entry
  for entry in "${servers[@]}"; do
    select_server "${entry%%|*}" || {
      rc=1
      continue
    }
    local server_rc=0
    case "$deploy_mode" in
    tmux) deploy_tmux "$skip_build" || server_rc=1 ;;
    systemctl) deploy_systemctl "$skip_build" "$force_service" || server_rc=1 ;;
    *) error "Unknown deploy_mode '$deploy_mode' (tmux|systemctl)" && return 1 ;;
    esac
    # Web-server config is installed by the deploy functions themselves, before
    # the backend is restarted; a failed server never reaches that point.
    [ $server_rc -eq 0 ] || rc=1
  done
  [ $rc -eq 0 ] && success "Deployed to all servers."
  return $rc
}

# Refuse to reuse artifacts that were built for a different server or a
# different API/site origin. Runs only with --skip-build, where nothing else
# re-checks the placeholder guard.
check_build_marker() {
  local entry name path cmd out_dir api site built_server="" built_api="" built_site="" k v rc=0
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    local marker="$path/$out_dir/.built-for"
    if [ ! -f "$marker" ]; then
      error "$name artifacts have no build marker (never built via main.sh); rebuild without --skip-build"
      rc=1
      continue
    fi
    built_server=""
    built_api=""
    built_site=""
    while IFS='=' read -r k v; do
      case "$k" in
      server) built_server="$v" ;;
      NUXT_PUBLIC_API_BASE) built_api="$v" ;;
      NUXT_PUBLIC_SITE_URL) built_site="$v" ;;
      esac
    done <"$marker"
    local -a envs=()
    app_envs "$name" envs
    api=$(env_value NUXT_PUBLIC_API_BASE envs || file_env "$path" NUXT_PUBLIC_API_BASE)
    site=$(env_value NUXT_PUBLIC_SITE_URL envs || file_env "$path" NUXT_PUBLIC_SITE_URL)
    if [ "$built_server" != "$srv_name" ]; then
      error "$name artifacts were built for '$built_server', target is '$srv_name'"
      rc=1
    fi
    if [ "$built_api" != "$api" ] || [ "$built_site" != "$site" ]; then
      error "$name artifacts were built with API='$built_api' site='$built_site'; current config wants API='$api' site='$site'"
      rc=1
    fi
    check_public_env "$name" deploy "$api" "$site" || rc=1
  done
  return $rc
}

# Prepare build artifacts and zip them for deployment.
prepare_deploy() {
  local skip_build="$1"
  if [ "$skip_build" = false ]; then
    build_front || {
      error "Frontend build failed, aborting deploy."
      return 1
    }
  else
    check_build_marker
  fi
  [ -f bin/backend ] || {
    error "bin/backend missing — run ./main.sh build"
    return 1
  }
  command -v zip >/dev/null 2>&1 || {
    error "zip not installed"
    return 1
  }

  local -a files=(bin/backend)
  local entry name path cmd out_dir site_d
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ -e "$path/$out_dir" ] || {
      error "Missing build artifact: $path/$out_dir"
      return 1
    }
    if [ "$web_server" != none ]; then
      site_d=$(domain_of "$name") || site_d=""
      if [ -n "$site_d" ] && [ ! -d "$path/$out_dir/public" ]; then
        error "app '$name' has domain '$site_d' but $path/$out_dir/public is missing (not an SSG build?)"
        return 1
      fi
    fi
    files+=("$path/$out_dir")
  done

  # package
  ensure_tmp_dir || return 1
  rm -f "$tmp_dir/web.zip"
  info "Packaging..."
  zip -r "$tmp_dir/web.zip" "${files[@]}" || {
    error "Packaging failed, aborting deploy."
    return 1
  }
}

deploy_tmux() {
  local skip_build="$1"
  prepare_deploy "$skip_build" || return 1
  # remote
  remote "mkdir -p $tmux_dir" || return 1
  sync_conf_if_missing "$tmux_dir" || return 1
  remote_util "$tmp_dir/web.zip" "$tmux_dir" || return 1
  rm -f "$tmp_dir/web.zip"
  # web config goes in before the backend restarts, so the process never comes
  # up behind a site config that still points at the old tree
  install_webconf || return 1
  patch_conf_toml || return 1

  # remote shell
  info "Restarting remote service..."
  remote "bash -s" -- "$tmux_dir" "$pname" <<'EOF' || return 1
        set -e
        REMOTE_DIR="$1"
        REMOTE_PNAME="$2"
        cd "$REMOTE_DIR" || { echo "Directory not found" >&2; exit 1; }
        command -v unzip >/dev/null 2>&1 || { echo "unzip missing on remote" >&2; exit 1; }
        unzip -tq web.zip >/dev/null || { echo "web.zip is corrupt" >&2; exit 1; }
        tmux kill-session -t "$REMOTE_PNAME" 2>/dev/null || true
        tmux new-session -d -s "$REMOTE_PNAME" -n "server"
        tmux new-window -t "$REMOTE_PNAME:1" -n "panel"
        tmux send-keys -t "$REMOTE_PNAME:0" "unzip -qo web.zip && rm web.zip && chmod +x ./bin/backend && ./bin/backend" C-m
EOF
  # The tmux session exists as soon as send-keys returns, so poll for the
  # process itself; a crash-on-start would otherwise be reported as success.
  if remote "command -v pgrep >/dev/null 2>&1"; then
    local up=false
    for _ in $(seq 1 15); do
      if remote "pgrep -f '$tmux_dir/bin/backend' >/dev/null 2>&1"; then
        up=true
        break
      fi
      sleep 1
    done
    if [ "$up" != true ]; then
      error "backend process did not appear in tmux session $pname"
      return 1
    fi
  else
    warn "pgrep missing on remote, skipping the backend process check."
  fi
  info "Tmux deployment successful."
}

# The deploy dir holds conf.toml (secrets) and the built static tree. A web
# server on the same host must traverse it to serve /public and /media, so it
# stays 0755 when web_server != none; without a web server nothing outside the
# service needs access, so it is 0750.
tighten_deploy_dir() {
  local d="$1" mode=0755
  [ "$web_server" = "none" ] && mode=0750
  remote "chmod $mode $d" || return 1
}

sync_conf_if_missing() {
  local remote_dir="$1"
  if ! remote "test -f $remote_dir/conf.toml"; then
    info "No remote conf.toml, syncing local conf.toml..."
    remote_util conf.toml "$remote_dir" || return 1
    # conf.toml holds the JWT secrets: it is never world-readable.
    remote "chmod 600 $remote_dir/conf.toml" || return 1
  else
    info "Remote conf.toml exists, skipping sync to avoid overwrite."
  fi
  tighten_deploy_dir "$remote_dir"
}

# Resolve the absolute paths of the privileged binaries on the remote host.
# sudoers entries must be exact paths, and nginx/caddy live under different
# prefixes across distributions, so nothing here is hard-coded.
resolve_remote_bins() {
  [ -n "$remote_systemctl" ] && return 0
  local resolved n p
  # Single quotes are intentional: the loop must expand on the remote host.
  # shellcheck disable=SC2016
  resolved=$(remote 'for b in systemctl install journalctl nginx caddy tee mkdir; do printf "%s=%s\n" "$b" "$(command -v "$b" 2>/dev/null)"; done') || return 1
  while IFS='=' read -r n p; do
    case "$n" in
    systemctl) remote_systemctl="$p" ;;
    install) remote_install="$p" ;;
    journalctl) remote_journalctl="$p" ;;
    nginx) remote_nginx="$p" ;;
    caddy) remote_caddy="$p" ;;
    tee) remote_tee="$p" ;;
    mkdir) remote_mkdir="$p" ;;
    esac
  done <<<"$resolved"
  [ -n "$remote_systemctl" ] && [ -n "$remote_install" ] || {
    error "cannot resolve systemctl/install on $ssh_target"
    return 1
  }
  return 0
}

# Ensure remote has limited passwordless sudo for the commands this deploy needs.
# All projects on the same server share one file: /etc/sudoers.d/<user>-<sudoers_name>.
# Idempotent: probes one whitelisted command and skips when it already works.
#
# NOTE: `install` writing into /etc and `systemctl` over units is, in practice,
# equivalent to root. That is unavoidable without a root-owned wrapper script, so
# use a dedicated deploy account on a machine you control (or deploy as root,
# which skips this entirely). The rules are still exact: no wildcards, and every
# path names this project.
ensure_sudoers() {
  resolve_remote_bins || return 1
  [ -n "$SUDO" ] || {
    info "srv_user=root, skipping sudoers setup."
    return 0
  }
  local sudoer_file="$srv_user-$sudoers_name" d
  d=$(deploy_dir) || return 1

  local -a cmds=(
    "$remote_systemctl enable $pname"
    "$remote_systemctl restart $pname"
    "$remote_systemctl stop $pname"
    "$remote_systemctl is-active $pname"
    "$remote_systemctl daemon-reload"
    "$remote_install -d -o root -g root -m 0755 $systemd_root"
    "$remote_install -o root -g root -m 0644 $d/$pname.service $sys/$pname.service"
  )
  if [ -n "$remote_nginx" ]; then
    cmds+=(
      "$remote_nginx -t"
      "$remote_nginx -v"
      "$remote_systemctl reload nginx"
      "$remote_install -o root -g root -m 0644 $d/.webconf /etc/nginx/conf.d/$pname.conf"
    )
  fi
  if [ -n "$remote_caddy" ]; then
    cmds+=(
      "$remote_caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile"
      "$remote_caddy version"
      "$remote_systemctl reload caddy"
      "$remote_mkdir -p /etc/caddy/conf.d"
      "$remote_install -o root -g root -m 0644 $d/.webconf /etc/caddy/conf.d/$pname.conf"
      "$remote_tee -a /etc/caddy/Caddyfile"
    )
  fi
  if [ -n "$remote_journalctl" ]; then
    cmds+=("$remote_journalctl -u $pname")
  fi
  local commands
  commands=$(printf '%s, ' "${cmds[@]}")
  commands="${commands%, }"
  local rule="$srv_user ALL=(root) NOPASSWD: $commands"

  if remote "sudo -n $remote_systemctl daemon-reload >/dev/null 2>&1"; then
    info "sudoers for $srv_user already configured, skipping."
    return 0
  fi
  if [ ! -t 0 ] && [ ! -t 1 ]; then
    error "First-time sudoers setup needs an interactive terminal. Run deploy once from a TTY, or install /etc/sudoers.d/$sudoer_file manually."
    return 1
  fi
  warn "Configuring limited passwordless sudo for $srv_user ($sudoer_file)."
  # The candidate is written to a private mktemp in the user's home, validated,
  # then installed with root ownership: never a predictable /tmp path.
  local rc=0
  ssh -t "$ssh_target" bash -s -- "$sudoer_file" "$rule" <<'EOF' || rc=$?
    set -e
    FILE="$1"; RULE="$2"
    umask 077
    TMP=$(mktemp "$HOME/.${FILE}.XXXXXX")
    printf '%s\n' "$RULE" > "$TMP"
    sudo visudo -cf "$TMP"
    sudo install -o root -g root -m 0440 "$TMP" "/etc/sudoers.d/$FILE"
    rm -f "$TMP"
EOF
  if [ "$rc" -ne 0 ]; then
    error "sudoers install failed (visudo rejected it or sudo was cancelled)"
    return 1
  fi
  info "sudoers configured."
}

# Full systemd deploy: sync build into /srv/http, install service, start.
# Requires limited passwordless sudo for the exact commands in ensure_sudoers.
deploy_systemctl() {
  local skip_build="$1" force_service="$2"
  prepare_deploy "$skip_build" || return 1
  resolve_remote_bins || return 1
  ensure_sudoers || return 1

  # a. render the unit; @RUN_USER@ is the deploy account, so the backend never
  #    runs as root
  sed -e "s/@PNAME@/$pname/g" -e "s|@SRV_DIR@|$systemd_dir|g" -e "s/@RUN_USER@/$srv_user/g" \
    assets/run.service >"$tmp_dir/$pname.service" || return 1

  # b. first-time dirs: a root-owned 0755 parent (never world-writable staging),
  #    then a project dir owned by the deploy user.
  if ! remote "test -d $systemd_root"; then
    info "First-time setup: creating $systemd_root..."
    remote "$SUDO $remote_install -d -o root -g root -m 0755 $systemd_root" || return 1
  fi
  if ! remote "test -d $systemd_dir"; then
    info "First-time setup: creating $systemd_dir..."
    remote "$SUDO $remote_install -d -o $srv_user -m 0750 $systemd_dir" || return 1
  fi

  # c. push the build into the deploy dir (owned by the deploy user, no sudo)
  info "Syncing to $systemd_dir..."
  remote_util "$tmp_dir/web.zip" "$systemd_dir" || return 1
  remote_util "$tmp_dir/$pname.service" "$systemd_dir" || return 1
  sync_conf_if_missing "$systemd_dir" || return 1

  # d. web config + conf.toml flags first, so the restarted backend reads the
  #    final config and the site never points at a half-updated tree
  install_webconf || return 1
  patch_conf_toml || return 1

  # e. extract + install service + start + health check (sudo whitelisted)
  info "Extracting & registering systemd service..."
  remote "bash -s" -- "$systemd_dir" "$sys" "$pname" "$force_service" "$SUDO" "$remote_systemctl" "$remote_install" <<'EOF' || return 1
        set -e
        SRV="$1"; SYS="$2"; P="$3"; FORCE="$4"; SUDO="$5"; SYSTEMCTL="$6"; INSTALL="$7"
        cd "$SRV" || { echo "Directory not found" >&2; exit 1; }
        [ -f web.zip ] || { echo "web.zip missing" >&2; exit 1; }
        command -v unzip >/dev/null 2>&1 || { echo "unzip missing on remote" >&2; exit 1; }
        unzip -tq web.zip >/dev/null
        # Stop before unpacking: replacing a running executable fails with ETXTBSY.
        $SUDO "$SYSTEMCTL" stop "$P" 2>/dev/null || true
        unzip -qo web.zip
        rm -f web.zip
        chmod +x bin/backend
        # Install whenever the rendered unit differs from the installed one.
        if [ "$FORCE" = "true" ] || [ ! -f "$SYS/${P}.service" ] || ! cmp -s "${SRV}/${P}.service" "$SYS/${P}.service"; then
          $SUDO "$INSTALL" -o root -g root -m 0644 "${SRV}/${P}.service" "$SYS/${P}.service"
          $SUDO "$SYSTEMCTL" daemon-reload
        fi
        rm -f "${SRV}/${P}.service"
        $SUDO "$SYSTEMCTL" enable "$P"
        $SUDO "$SYSTEMCTL" restart "$P"
        for i in $(seq 1 15); do
          $SUDO "$SYSTEMCTL" is-active "$P" >/dev/null 2>&1 && exit 0
          sleep 1
        done
        echo "Service $P failed to start" >&2
        exit 1
EOF
  info "Systemctl deployment successful."
}

# Local expression of the remote deploy dir for the current deploy_mode (systemctl|tmux).
deploy_dir() {
  case "$deploy_mode" in
  systemctl) printf '%s' "$systemd_dir" ;;
  tmux) printf '%s' "$tmux_dir" ;;
  *) error "Unknown deploy_mode '$deploy_mode' (tmux|systemctl)" && return 1 ;;
  esac
}

synconf() {
  local server="$2" d
  select_server "$server" || return 1
  d=$(deploy_dir) || return 1
  remote "mkdir -p $d" || return 1
  remote_util conf.toml "$d" || return 1
  remote "chmod 600 $d/conf.toml" || return 1
  tighten_deploy_dir "$d" || return 1
  info "conf.toml synced to $srv_name:$d"
}

# Copy an env template, keeping an existing file (it may hold local overrides)
# unless --force was passed.
# usage: copy_env <src> <dest> <true|false>
copy_env() {
  local src="$1" dest="$2" force="$3"
  if [ -f "$dest" ] && [ "$force" != true ]; then
    info "$dest exists, keeping it (use 'copyfile --force' to overwrite)."
    return 0
  fi
  cp "$src" "$dest" || return 1
}

copyfile() {
  local force=false arg
  for arg in "$@"; do
    case "$arg" in
    --force) force=true ;;
    esac
  done

  if [ -f ./assets/conf.toml ]; then
    # conf.toml holds the jwt secrets and every credential. Overwriting it with
    # the template wipes them and the app refuses to start, so only write when
    # the file is absent (or --force was passed).
    if [ -f ./conf.toml ] && [ "$force" = false ]; then
      warn "conf.toml already exists, keeping it (use 'copyfile --force' to overwrite)."
    else
      cp ./assets/conf.toml ./conf.toml
      info "Wrote conf.toml from template — fill in jwt.secret.adm / jwt.secret.usr."
    fi
  else
    info "No assets/conf.toml template, skipping (create conf.toml manually)."
  fi

  local entry name path cmd out_dir dev_env prod_env
  dev_env=./assets/nuxt.env
  prod_env=./assets/nuxt.env.production
  [ -f "$prod_env" ] || warn "$prod_env missing, .env.production will copy from the dev template."
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    # An existing .env/.env.production may hold local overrides; never clobber
    # it silently (this used to reset the dev env on every `copyfile`).
    copy_env "$dev_env" "$path/.env" "$force"
    # Production gets its own template so a dev build's localhost API base can
    # never leak into a deployed bundle.
    if [ -f "$prod_env" ]; then
      copy_env "$prod_env" "$path/.env.production" "$force"
    else
      copy_env "$dev_env" "$path/.env.production" "$force"
    fi
  done
}

init() {
  command -v gowatch >/dev/null 2>&1 || go install github.com/silenceper/gowatch@latest
  command -v gowatch >/dev/null 2>&1 || {
    error "gowatch not on PATH after install (check GOBIN)"
    return 1
  }
  if ! command -v swag >/dev/null 2>&1; then
    info "Installing swag..."
    go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc6 || {
      error "swag install failed."
      return 1
    }
    command -v swag >/dev/null 2>&1 || {
      error "swag not found on PATH after install (check GOBIN)."
      return 1
    }
  fi
  git submodule update --init --recursive || return 1
  mkdir -p bin
  # pnpm install and the submodule must be in place before docs: swag init walks
  # the workspace and gen:api runs through pnpm.
  run_in_dir web pnpm install || return 1
  go mod tidy || {
    error "go mod tidy failed"
    return 1
  }
  docs || return 1
}

docs() {
  info "Generating Swagger documentation"
  gen_swagger || return 1
  info "Regenerating frontend SDK"
  gen_sdk || return 1
}

# Lint + format gate for the Go backend and the deploy scripts.
#
# Formatting is checked with the standalone gofumpt/gci binaries because
# golangci-lint 2.14's built-in gofumpt formatter is unreliable (see
# .golangci.yml). Every tool is required: a missing formatter used to print an
# empty diff and report success, which is worse than failing loudly.
#
# Usage: main.sh lint          (check)
#        main.sh lint --fix    (apply formatting fixes first)
lint() {
  local fix=false arg
  for arg in "$@"; do
    case "$arg" in
    --fix) fix=true ;;
    -*)
      error "Unknown lint option: $arg"
      usage >&2
      return 1
      ;;
    *)
      error "Unexpected lint argument: $arg"
      usage >&2
      return 1
      ;;
    esac
  done

  local gobin="${GOBIN:-$HOME/.local/share/go/bin}"
  local failed=0 missing=0 tool unformatted
  info "Lint: tool preflight"
  for tool in gofumpt gci staticcheck golangci-lint; do
    if [ ! -x "$gobin/$tool" ]; then
      error "missing $gobin/$tool"
      missing=1
    fi
  done
  for tool in go shellcheck; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      error "missing $tool on PATH"
      missing=1
    fi
  done
  if [ "$missing" -ne 0 ]; then
    error "ABORTING: install the missing tools (see AGENTS.md -> Lint) and re-run"
    return 1
  fi
  success "all tools present"

  if [ "$fix" = true ]; then
    info "Lint: formatting (gofumpt + gci)"
    "$gobin/gofumpt" -w . || return 1
    "$gobin/gci" write -s standard -s default -s 'Prefix(webgx)' . || return 1
  fi

  info "Lint: formatting check"
  unformatted=$("$gobin/gofumpt" -l . 2>/dev/null | grep -v '^$' || true)
  if [ -n "$unformatted" ]; then
    error "gofumpt wants to rewrite:"
    echo "$unformatted"
    error "run: ./main.sh lint --fix"
    failed=1
  else
    success "gofumpt: clean"
  fi

  if [ -n "$("$gobin/gci" diff -s standard -s default -s 'Prefix(webgx)' . 2>/dev/null || true)" ]; then
    error "gci wants to reorder imports (run ./main.sh lint --fix)"
    failed=1
  else
    success "gci: clean"
  fi

  info "Lint: go build"
  go build ./... || {
    error "build failed"
    failed=1
  }

  info "Lint: go vet"
  go vet ./... || {
    error "vet failed"
    failed=1
  }

  info "Lint: staticcheck"
  "$gobin/staticcheck" ./... || {
    error "staticcheck failed"
    failed=1
  }

  info "Lint: golangci-lint"
  "$gobin/golangci-lint" run ./... || {
    error "golangci-lint failed"
    failed=1
  }

  info "Lint: shellcheck (deploy scripts)"
  shellcheck -x main.sh || {
    error "shellcheck failed"
    failed=1
  }

  echo
  if [ "$failed" -ne 0 ]; then
    error "LINT FAILED"
    return 1
  fi
  success "ALL CHECKS PASSED"
}

renewal() {
  local new_name="" skip_backup=false arg SED_CMD=()
  for arg in "$@"; do
    case "$arg" in
    --no-backup) skip_backup=true ;;
    -*)
      error "Unknown renewal option: $arg"
      usage >&2
      return 1
      ;;
    *)
      if [ -n "$new_name" ]; then
        error "Unexpected renewal argument: $arg"
        return 1
      fi
      new_name="$arg"
      ;;
    esac
  done
  if [ -z "$new_name" ]; then
    read -rp "Input new project name: " new_name
  fi
  if [ -z "$new_name" ]; then
    warn "Missing Project Name."
    return 0
  fi
  validate_pname "$new_name" || return 1
  if [ "$new_name" = "$pname" ]; then
    warn "New name '$new_name' equals the current name; nothing to do."
    return 0
  fi
  # renewal deletes .git before touching the tree, so uncommitted work would be
  # unrecoverable: refuse to start unless the worktree is clean.
  if [ -d .git ] && [ -n "$(git status --porcelain)" ]; then
    error "Working tree is dirty; commit or stash your changes before renewal."
    return 1
  fi

  echo -e "Project Initialization & Reset"
  echo "--------------------------------------------------"
  echo "This script will perform the following IRREVERSIBLE actions:"
  echo -e "1. GLOBAL REPLACE: All occurrences of '$pname' (including main.sh) will be changed to '$new_name'."
  echo -e "2. GIT RESET: The existing .git directory will be DELETED."
  if [ "$skip_backup" = true ]; then
    echo -e "   (--no-backup: no git bundle, no conf.toml copy — history and secrets are LOST)"
  else
    echo -e "   (history is bundled next to the repo first)"
  fi
  echo -e "3. RE-INIT: A new git repository will be initialized in this directory."
  echo -e "4. CONFIG: .env files will be recreated from assets/; conf.toml is kept, including its [db] name."
  echo "--------------------------------------------------"

  # prompt for user confirmation
  read -rp "Are you sure you want to proceed? (y/N): " confirm
  if [[ "$confirm" =~ ^[yY](es)?$ ]]; then
    local backup_prefix db_name=""
    backup_prefix="$(dirname "$PWD")/${pname}-pre-renewal-$(date +%Y%m%d-%H%M%S)"
    if [ "$skip_backup" = true ]; then
      warn "Skipping backups (--no-backup): git history and conf.toml are unrecoverable, and [db] name follows the rename."
    else
      if git rev-parse --verify HEAD >/dev/null 2>&1; then
        if ! git bundle create "${backup_prefix}.bundle" --all >/dev/null; then
          error "git bundle failed; aborting before .git is deleted."
          return 1
        fi
        info "Git history backed up to ${backup_prefix}.bundle"
      else
        warn "No commits to bundle; skipping git history backup."
      fi
      # The live database is named after the old project; remember it so the
      # rename does not repoint the app at a database that does not exist.
      if [ -f conf.toml ]; then
        db_name=$(sed -n '/^\[db\]/,/^\[/ s/^[[:space:]]*name[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' conf.toml | head -n1)
        if [ -n "$db_name" ] && [[ ! "$db_name" =~ ^[A-Za-z0-9_$-]+$ ]]; then
          warn "Unusual database name '$db_name'; leaving [db] name to the rename."
          db_name=""
        fi
        cp conf.toml "${backup_prefix}.conf.toml" || return 1
        info "conf.toml backed up to ${backup_prefix}.conf.toml"
      fi
    fi
    # reset git directory
    rm -rf .git
    rm -rf web/packages/natholdallas
    rm .gitmodules
    # replace all occurrences of old_name with new_name (escape sed special chars)
    if [[ "$OSTYPE" == "darwin"* ]]; then
      # macOS syntax
      SED_CMD=(sed -i "")
    else
      # Linux syntax
      SED_CMD=(sed -i)
    fi
    local pname_esc new_name_esc
    pname_esc=$(sed 's/[\/&\\]/\\&/g' <<<"$pname")
    new_name_esc=$(sed 's/[\/&\\]/\\&/g' <<<"$new_name")
    find . \
      \( -name ".git" \
      -o -name "node_modules" \
      -o -name ".nuxt" \
      -o -name ".output" \
      -o -name "dist" \
      -o -name "docs" \
      -o -name "logs" \
      -o -name "media" \
      -o -name "bin" \
      -o -name "go.sum" \
      -o -name "pnpm-lock.yaml" \
      -o -name "target" \
      -o -name "build" \
      -o -name ".gradle" \
      -o -name ".kotlin" \
      -o -name "*.pem" \) -prune \
      -o -type f \
      \( -name "*.go" -o -name "*.md" -o -name "*.mod" -o -name "*.toml" \
      -o -name "*.service" -o -name "*.sh" -o -name "*.json" \
      -o -name "*.yml" -o -name "*.yaml" -o -name "*.ts" \
      -o -name "*.tsx" -o -name "*.vue" -o -name "*.nuxtrc" \
      -o -name "*.conf" -o -name "*.rs" -o -name "*.kt" -o -name "*.kts" \
      -o -name "*.xml" -o -name "*.pro" -o -name "*.css" -o -name "*.scss" \
      -o -name "*.html" -o -name "*.example" -o -name "Caddyfile*" \
      -o -name "*.env" -o -name "*.lock" -o -name ".gitmodules" \) \
      -exec "${SED_CMD[@]}" "s/${pname_esc}/${new_name_esc}/g" {} +
    # sed rewrites file contents only; directories named after the project (the
    # Android package path com/<pname>/...) need an explicit rename, deepest first.
    find . \
      \( -name ".git" -o -name "node_modules" -o -name "target" \
      -o -name ".gradle" -o -name "build" -o -name ".nuxt" \
      -o -name ".output" -o -name "dist" \) -prune \
      -o -type d -name "*${pname}*" -print | sort -r |
      while IFS= read -r d; do
        local base="${d##*/}"
        mv "$d" "${d%/*}/${base//"$pname"/"$new_name"}"
      done
    # sed pointed [db] name at the new project; restore it so the existing data
    # (a database named after the old project) stays reachable.
    if [ -n "$db_name" ]; then
      "${SED_CMD[@]}" '/^\[db\]/,/^\[/ s/^\([[:space:]]*name[[:space:]]*=[[:space:]]*\).*/\1"'"$db_name"'"/' conf.toml
      info "conf.toml [db] name kept as '$db_name'"
    fi
    # initialize git repository
    git init || return 1
    git submodule add https://github.com/natholdallas/nuxt-modules.git web/packages/natholdallas || {
      error "git submodule add failed; re-run it inside web/packages/natholdallas' parent."
      return 1
    }
    # copy file
    copyfile || return 1
    # initialize project
    init || {
      error "Project init failed; fix the reported step and re-run ./main.sh init."
      return 1
    }
    # add all files & commit
    git add -A
    git commit -m "initialize"
    success "Project initialized successfully."
  fi
}

cleanup() {
  [ -n "$tmp_dir" ] && [ -d "$tmp_dir" ] || return 0
  rm -rf "$tmp_dir"
  info "Cleaned up $tmp_dir"
  tmp_dir=""
}

on_interrupt() {
  if [ "${current_cmd:-}" = dev ] && tmux has-session -t "$pname" 2>/dev/null; then
    warn "Stopping tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
  fi
  exit 1
}

serverlog() {
  local server="" arg="$2"
  if [ -n "$arg" ] && is_server "$arg"; then
    server="$arg"
    arg="$3"
  fi
  select_server "$server" || return 1
  # `once` is a mode, not a line count: resolve it before defaulting `lines`,
  # otherwise journalctl received "-n once" and errored out.
  local follow=true lines=100
  case "$arg" in
  once) follow=false ;;
  "") : ;;
  *)
    if [[ "$arg" =~ ^[0-9]+$ ]]; then
      lines="$arg"
    else
      warn "Unknown serverlog argument '$arg' (expected 'once' or a line count)."
    fi
    ;;
  esac
  if [ "$follow" = true ]; then
    ssh -t "$ssh_target" "$SUDO journalctl -u $pname -n $lines -f"
  else
    ssh -t "$ssh_target" "$SUDO journalctl -u $pname -n $lines"
  fi
}

clean() {
  find . -name "node_modules" -type d -prune -exec rm -rf {} +
  find . -name "dist" -type d -prune -exec rm -rf {} +
  find . -name ".nuxt" -type d -prune -exec rm -rf {} +
  find . -name ".output" -type d -prune -exec rm -rf {} +
}

push() {
  local commit_info=$2
  if [ -z "$2" ]; then
    warn "Missing commit message, using default 'upd'"
    commit_info=upd
  fi
  git add -A
  if ! git commit -m "$commit_info"; then
    error "Commit failed, aborting push."
    return 1
  fi
  local branch
  branch=$(git branch --show-current 2>/dev/null)
  if [ -z "$branch" ]; then
    warn "Detached HEAD — cannot determine upstream."
  elif ! git config "branch.$branch.remote" >/dev/null 2>&1 && ! git branch -r | grep -q "origin/$branch$"; then
    warn "No upstream for branch '$branch'. Set upstream: git push -u origin $branch"
  fi
  git push
}

# Create a GitHub repo, bind origin, set upstream (never pushes). Usage: README.
initremote() {
  local visibility=private name="$pname"
  for arg in "$@"; do
    case "$arg" in
    --public) visibility=public ;;
    --private) visibility=private ;;
    *) name="$arg" ;;
    esac
  done

  # preflight: gh CLI must be installed and authenticated
  if ! command -v gh >/dev/null 2>&1; then
    error "gh CLI not found, please install it first: https://cli.github.com"
    return 1
  fi
  if ! gh auth status >/dev/null 2>&1; then
    error "gh not authenticated, please run: gh auth login"
    return 1
  fi

  # remember whether origin already exists, to decide bind vs rebind below
  local had_origin=false
  if git remote get-url origin >/dev/null 2>&1; then
    had_origin=true
    warn "Remote origin already exists: $(git remote get-url origin)"
  fi

  # step 1: create the repo on GitHub (no push, no README)
  info "Creating GitHub repo '$name' (--$visibility)..."
  gh repo create "$name" --"$visibility" || return 1

  # step 2: bind the new repo as origin
  local owner_repo proto want_url current_url
  owner_repo=$(gh repo view "$name" --json nameWithOwner -q .nameWithOwner 2>/dev/null)
  if [ -z "$owner_repo" ]; then
    error "Could not resolve repo '$name'."
    return 1
  fi
  # build the repo url according to gh's configured protocol (ssh/https)
  proto=$(gh config get git_protocol 2>/dev/null || echo https)
  if [ "$proto" = "ssh" ]; then
    want_url="git@github.com:$owner_repo.git"
  else
    want_url="https://github.com/$owner_repo.git"
  fi
  if [ "$had_origin" = true ]; then
    current_url=$(git remote get-url origin)
    if [ "$current_url" != "$want_url" ]; then
      warn "Rebinding origin: $current_url -> $want_url"
      git remote set-url origin "$want_url" || return 1
    else
      info "Origin already points at $want_url"
    fi
  else
    info "Adding remote origin -> $want_url"
    git remote add origin "$want_url" || return 1
  fi

  # step 3: set upstream (equivalent to push -u) WITHOUT pushing
  local branch
  branch=$(git branch --show-current 2>/dev/null)
  if [ -n "$branch" ]; then
    git config "branch.$branch.remote" origin || return 1
    git config "branch.$branch.merge" "refs/heads/$branch" || return 1
    success "Upstream set: $branch -> origin/$branch (nothing pushed)"
  else
    warn "Detached HEAD — skipping upstream setup."
  fi

  success "Repo: $(gh repo view "$name" --json url -q .url 2>/dev/null)"
  warn "Push later with: ./main.sh push \"commit message\""
}

# =============================================================================
# Web server auto-configuration (nginx / caddy)
# =============================================================================

# Every site endpoint: each app name (its site), then api and media.
# Adding an app to apps[] automatically adds its endpoint here.
endpoints() {
  local entry nm
  for entry in "${apps[@]}"; do
    IFS='|' read -r nm _ <<<"$entry"
    printf '%s\n' "$nm"
  done
  printf '%s\n' api media
}

# Print the domain for an endpoint (empty when unset): domain_<endpoint>.
domain_of() {
  local var="domain_$1"
  printf '%s' "${!var:-}"
  return 0
}

# Print configured domains as "endpoint=domain" list (for success messages).
get_domain_list() {
  local ep d out=()
  while IFS= read -r ep; do
    d=$(domain_of "$ep")
    [ -n "$d" ] && out+=("$ep=$d")
  done < <(endpoints)
  printf '%s' "${out[*]}"
}

# Append auto-derived env vars for an app (NUXT_PUBLIC_API_BASE / SITE_URL):
# the site endpoint is the app's own name, the API endpoint is always `api`.
# Explicit entries already in the envs array win.
# usage: auto_envs <app> <envs_array_name>
auto_envs() {
  local app="$1" api_d site_d v found
  local -n _envs="$2"
  api_d=$(domain_of api) || api_d=""
  site_d=$(domain_of "$app") || site_d=""
  if [ -n "$api_d" ]; then
    found=0
    for v in "${_envs[@]}"; do [ "${v%%=*}" = "NUXT_PUBLIC_API_BASE" ] && found=1; done
    [ $found -eq 0 ] && _envs+=("NUXT_PUBLIC_API_BASE=https://$api_d")
  fi
  if [ -n "$site_d" ]; then
    found=0
    for v in "${_envs[@]}"; do [ "${v%%=*}" = "NUXT_PUBLIC_SITE_URL" ] && found=1; done
    [ $found -eq 0 ] && _envs+=("NUXT_PUBLIC_SITE_URL=https://$site_d")
  fi
  return 0
}

# Resolve the absolute deploy dir on the remote (nginx/caddy root needs absolutes).
remote_abs_dir() {
  if [ "$deploy_mode" = "tmux" ]; then
    remote "cd $tmux_dir && pwd" || return 1
  else
    deploy_dir
  fi
}

# Read a key from the [app] section of a conf.toml (comments/quotes stripped).
# usage: conf_app_get <file> <key>   — returns 1 when the file or key is absent.
conf_app_get() {
  local file="$1" key="$2"
  [ -f "$file" ] || return 1
  awk -v k="$key" '
    { line = $0; sub(/#.*/, "", line); gsub(/^[ \t]+|[ \t]+$/, "", line) }
    line == "" { next }
    line ~ /^\[/ { inapp = (line == "[app]"); next }
    !inapp { next }
    index(line, "=") == 0 { next }
    { key = line; sub(/[ \t]*=.*/, "", key); if (key != k) next
      val = line; sub(/^[^=]*=[ \t]*/, "", val); gsub(/"/, "", val); print val; exit }
  ' "$file"
}

# Copy the remote conf.toml into the packaging dir, so the rendered site config
# is derived from the file the service will actually read. A missing remote file
# leaves the local conf.toml in use.
fetch_remote_conf() {
  local remote_conf
  remote_conf="$(deploy_dir)/conf.toml"
  remote "test -f $remote_conf" || return 0
  ensure_tmp_dir || return 1
  if ! remote "cat $remote_conf" >"$tmp_dir/remote-conf.toml"; then
    rm -f "$tmp_dir/remote-conf.toml"
    warn "cannot read $remote_conf, falling back to the local conf.toml"
  fi
  return 0
}

# Derive the reverse-proxy upstream and the client body limit from conf.toml so
# they cannot drift from the backend config. Environment variables win, and the
# remote conf.toml is preferred over the local one when it is available.
resolve_webconf_vars() {
  local port bl src
  src="$tmp_dir/remote-conf.toml"
  [ -n "$tmp_dir" ] && [ -f "$src" ] || src="conf.toml"
  if [ -n "${WEB_API_ADDR:-}" ]; then
    web_api_addr="$WEB_API_ADDR"
  else
    port=$(conf_app_get "$src" port) || port=""
    web_api_addr="127.0.0.1:${port:-8080}"
  fi
  if [ -n "${CLIENT_MAX_BODY:-}" ]; then
    client_max_body="$CLIENT_MAX_BODY"
  else
    bl=$(conf_app_get "$src" body-limit) || bl=""
    case "$bl" in
    "")
      client_max_body="20m"
      ;;
    *[kKmMgG]) client_max_body="$bl" ;;
    *) client_max_body="${bl}m" ;;
    esac
  fi
  [[ $client_max_body =~ ^[0-9]+[kKmMgG]?$ ]] || {
    error "client_max_body '$client_max_body' must look like 20m"
    return 1
  }
  # Caddy spells the same limit with a byte unit suffix (20m -> 20MB).
  case "$client_max_body" in
  *[kK]) client_max_body_caddy="${client_max_body%[kK]}KB" ;;
  *[mM]) client_max_body_caddy="${client_max_body%[mM]}MB" ;;
  *[gG]) client_max_body_caddy="${client_max_body%[gG]}GB" ;;
  *) client_max_body_caddy="${client_max_body}MB" ;;
  esac
  return 0
}

# Built-in SPA site block, used for an app endpoint that has no
# assets/<kind>/<endpoint>.conf override. Adding an app + domain_<app> therefore
# needs no new template file. Mirrors assets/{nginx,caddy}/{usr,adm}.conf: the
# site proxies every backend surface (same-origin mode) and serves /media.
# usage: default_site <nginx|caddy> <domain> <public dir> <media dir>
default_site() {
  local kind="$1" domain="$2" dir="$3" media="$4"
  case "$kind" in
  nginx)
    cat <<EOF
# Auto-generated by main.sh ($domain) — no assets/nginx/<endpoint>.conf, default SPA block
server {
    listen 80;
    server_name $domain;
    root $dir;
    index index.html;
    client_max_body_size $client_max_body;

    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options DENY always;
    add_header Referrer-Policy no-referrer always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://telegram.org; connect-src 'self' https:; img-src 'self' data: blob: https:; media-src 'self' blob: https:; style-src 'self' 'unsafe-inline'; font-src 'self' data:; frame-src https://oauth.telegram.org https://telegram.org; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'" always;

    # hashed build assets can be cached forever
    location /_nuxt/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files \$uri =404;
    }

    # same-origin deployment (NUXT_PUBLIC_API_BASE=same-origin): every backend
    # surface, including the auth cookie path
    location ~ ^/(api|usr|adm)/ {
        proxy_pass http://$web_api_addr;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # media with no media domain (resources.media-url empty → relative /media/...)
    location /media/ {
        alias $media/;
        autoindex off;
        add_header X-Content-Type-Options nosniff always;
        add_header Referrer-Policy no-referrer always;
    }

    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
    ;;
  caddy)
    cat <<EOF
# Auto-generated by main.sh ($domain) — no assets/caddy/<endpoint>.conf, default SPA block
$domain {
    encode gzip zstd

    request_body {
        max_size $client_max_body_caddy
    }

    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        Referrer-Policy no-referrer
        Permissions-Policy "camera=(), microphone=(), geolocation=()"
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://telegram.org; connect-src 'self' https:; img-src 'self' data: blob: https:; media-src 'self' blob: https:; style-src 'self' 'unsafe-inline'; font-src 'self' data:; frame-src https://oauth.telegram.org https://telegram.org; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"
    }

    @static path /_nuxt/*
    header @static Cache-Control "public, max-age=31536000, immutable"

    # same-origin deployment (NUXT_PUBLIC_API_BASE=same-origin). Handles are
    # mutually exclusive and evaluated in file order, so they precede the SPA
    # fallback block at the end.
    handle /api/* {
        reverse_proxy $web_api_addr
    }
    handle /usr/* {
        reverse_proxy $web_api_addr
    }
    handle /adm/* {
        reverse_proxy $web_api_addr
    }

    # media with no media domain (resources.media-url empty → relative /media/...)
    handle_path /media/* {
        root * $media
        file_server
    }

    handle {
        root * $dir
        try_files {path} /index.html
        file_server
    }
}
EOF
    ;;
  *)
    error "Unknown web server kind '$kind' (nginx|caddy)"
    return 1
    ;;
  esac
}

# Compose the web server config into $tmp_dir/$pname.conf from assets/<kind>/*
# (or the built-in default for app endpoints), substituting domain + deploy-dir
# tokens: @DOMAIN@, @SRV_DIR@, @MEDIA_DIR@, @API_ADDR@, @CLIENT_MAX_BODY@ and one
# @<APP>_DIR@ per entry in apps[] (upper-cased app name).
# usage: compose_webconf <nginx|caddy> <abs_deploy_dir>   (returns 2 if all empty)
compose_webconf() {
  local kind="$1" abs_dir="$2"
  case "$kind" in nginx | caddy) ;; *)
    error "Unknown web server kind '$kind' (nginx|caddy)"
    return 1
    ;;
  esac
  resolve_webconf_vars || return 1
  local entry name path cmd out_dir ep domain fname conf_text=""
  local media_dir="$abs_dir/media"
  local -A app_dir=()
  local -a eps=() sedargs=() names=()
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    app_dir[$name]="$abs_dir/$path/$out_dir/public"
    names+=("$name")
  done
  mapfile -t eps < <(endpoints)
  for ep in "${eps[@]}"; do
    domain=$(domain_of "$ep")
    [ -n "$domain" ] || continue
    sedargs=(-e "s|@DOMAIN@|$domain|g" -e "s|@SRV_DIR@|$abs_dir|g"
      -e "s|@MEDIA_DIR@|$media_dir|g" -e "s|@API_ADDR@|$web_api_addr|g"
      -e "s|@CLIENT_MAX_BODY@|$client_max_body|g"
      -e "s|@CLIENT_MAX_BODY_CADDY@|$client_max_body_caddy|g")
    for name in "${names[@]}"; do
      sedargs+=(-e "s|@${name^^}_DIR@|${app_dir[$name]}|g")
    done
    fname="assets/$kind/$ep.conf"
    if [ -f "$fname" ]; then
      conf_text+=$(sed "${sedargs[@]}" "$fname")$'\n'
    elif [ -n "${app_dir[$ep]:-}" ]; then
      conf_text+=$(default_site "$kind" "$domain" "${app_dir[$ep]}" "$media_dir")$'\n'
    else
      error "Missing web template: $fname"
      return 1
    fi
  done
  [ -n "$conf_text" ] || {
    warn "no domain_* set in deploy.conf, skipping web config generation."
    return 2
  }
  ensure_tmp_dir || return 1
  printf '%s' "$conf_text" >"$tmp_dir/$pname.conf" || return 1
  return 0
}

# Make the deployed conf.toml production-correct. Runs on every deploy (and
# standalone via `webconf`), so a stale or append-only remote file cannot leave
# the backend in debug mode, exposing the docs route, or trusting old origins:
#   [app]       debug = false, swagger = false, nginx = true (when fronted)
#   [resources] media-url = "https://<media>" (only when domain_media is set)
#   [cors]      prd rebuilt (assets/patch-conf.awk)
# The remote file is fetched first, patched locally, then uploaded back, so the
# transform is one testable script instead of a remote quoting puzzle.
patch_conf_toml() {
  local d remote_conf media_d sites="" ep dd rc=0
  d=$(deploy_dir) || return 1
  remote_conf="$d/conf.toml"
  remote "test -f $remote_conf" || {
    warn "No remote conf.toml at $remote_conf, skipping patch."
    return 0
  }
  ensure_tmp_dir || return 1
  remote "cat $remote_conf" >"$tmp_dir/conf.toml" || return 1
  media_d=$(domain_of media) || media_d=""
  # Only site endpoints render a page, so only they belong in [cors] prd.
  while IFS= read -r ep; do
    case "$ep" in api | media) continue ;; esac
    dd=$(domain_of "$ep")
    [ -n "$dd" ] && sites+=" $dd"
  done < <(endpoints)
  local media_url="" nginx_mode=false
  [ -n "$media_d" ] && media_url="https://$media_d"
  [ "$web_server" != none ] && nginx_mode=true
  awk -v media_url="$media_url" -v sites="${sites# }" -v nginx_mode="$nginx_mode" \
    -f assets/patch-conf.awk "$tmp_dir/conf.toml" >"$tmp_dir/conf.toml.patched" || rc=$?
  if [ "$rc" -ne 0 ]; then
    error "conf.toml patch failed (see patch-conf above); fix $remote_conf and redeploy"
    return 1
  fi
  if cmp -s "$tmp_dir/conf.toml" "$tmp_dir/conf.toml.patched"; then
    info "conf.toml already production-correct (debug/swagger off, nginx flag, cors.prd)"
    return 0
  fi
  remote_util "$tmp_dir/conf.toml.patched" "$d" || return 1
  remote "mv $d/conf.toml.patched $d/conf.toml && chmod 600 $d/conf.toml" || return 1
  info "conf.toml patched (debug/swagger off, nginx flag, media-url, cors.prd)"
}

# Install & reload the web server site config for the given kind (nginx|caddy).
install_server() {
  local kind="$1" abs_dir="$2" bin rc d
  case "$kind" in
  nginx | caddy) bin="$kind" ;;
  *)
    error "Unknown web server kind '$kind' (nginx|caddy)"
    return 1
    ;;
  esac

  compose_webconf "$kind" "$abs_dir"
  rc=$?
  case "$rc" in
  0) ;;
  2) return 0 ;; # no domains configured, nothing to install
  *) return 1 ;;
  esac

  remote "command -v $bin >/dev/null 2>&1" || {
    error "$bin not installed on $ssh_target — run: sudo apt install $bin"
    return 1
  }
  d=$(deploy_dir) || return 1
  # Stage inside the deploy dir (owned by the deploy user): a predictable /tmp
  # path could be pre-created or swapped by another local account between write
  # and `sudo install`.
  remote_util "$tmp_dir/$pname.conf" "$d" || return 1
  remote "mv $d/$pname.conf $d/.webconf && chmod 600 $d/.webconf" || return 1
  remote "bash -s" -- "$d" "$pname" "$bin" "$SUDO" "$remote_systemctl" "$remote_install" \
    "$remote_nginx" "$remote_caddy" "$remote_mkdir" "$remote_tee" <<'EOF' || return 1
    set -e
    D="$1"; P="$2"; BIN="$3"; SUDO="$4"; SYSTEMCTL="$5"; INSTALL="$6"
    NGINX="$7"; CADDY="$8"; MKDIR="$9"; TEE="${10}"
    SRC="$D/.webconf"
    PREV="$D/.webconf.prev"
    [ -f "$SRC" ] || { echo "staged config missing: $SRC" >&2; exit 1; }

    # restore the previous file, or an empty one when this is the first install
    restore() {
      if [ -f "$1" ]; then
        $SUDO "$INSTALL" -o root -g root -m 0644 "$1" "$2"
      else
        $SUDO "$INSTALL" -o root -g root -m 0644 /dev/null "$2"
      fi
    }

    case "$BIN" in
    nginx)
      DEST="/etc/nginx/conf.d/${P}.conf"
      rm -f "$PREV"
      [ -f "$DEST" ] && cp "$DEST" "$PREV"
      $SUDO "$INSTALL" -o root -g root -m 0644 "$SRC" "$DEST"
      if $SUDO "$NGINX" -t >/dev/null 2>&1; then
        $SUDO "$SYSTEMCTL" reload nginx
      else
        restore "$PREV" "$DEST"
        if ! $SUDO "$NGINX" -t >/dev/null 2>&1; then
          echo "nginx config test FAILED and the restored config is invalid too; left it in place for inspection" >&2
          exit 1
        fi
        echo "nginx config test FAILED for ${P}.conf; previous config restored" >&2
        exit 1
      fi
      ;;
    caddy)
      DEST="/etc/caddy/conf.d/${P}.conf"
      $SUDO "$MKDIR" -p /etc/caddy/conf.d
      rm -f "$PREV"
      [ -f "$DEST" ] && cp "$DEST" "$PREV"
      $SUDO "$INSTALL" -o root -g root -m 0644 "$SRC" "$DEST"
      # The import line is a one-time edit of the main Caddyfile. Without it the
      # site block is never loaded, so a failed check is an error, not a warning.
      if ! grep -qs 'import conf.d/\*\.conf' /etc/caddy/Caddyfile; then
        if [ ! -r /etc/caddy/Caddyfile ]; then
          echo "cannot read /etc/caddy/Caddyfile; add 'import conf.d/*.conf' to it once, then redeploy" >&2
          exit 1
        fi
        printf '\nimport conf.d/*.conf\n' | $SUDO "$TEE" -a /etc/caddy/Caddyfile >/dev/null
      fi
      if $SUDO "$CADDY" validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null 2>&1; then
        $SUDO "$SYSTEMCTL" reload caddy
      else
        restore "$PREV" "$DEST"
        if ! $SUDO "$CADDY" validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null 2>&1; then
          echo "caddy validate FAILED and the restored config is invalid too; left it in place for inspection" >&2
          exit 1
        fi
        echo "caddy validate FAILED for ${P}.conf; previous config restored" >&2
        exit 1
      fi
      ;;
    esac
    rm -f "$SRC" "$PREV"
EOF
  success "$bin configured & reloaded (domains: $(get_domain_list))"
}

# Install the web server site config for the current server (deploy-mode aware).
# Called after each deploy, and standalone via: ./main.sh webconf [server]
install_webconf() {
  [ "$web_server" = "none" ] && {
    info "web_server=none, skipping web config."
    return 0
  }
  local abs_dir
  resolve_remote_bins || return 1
  abs_dir=$(remote_abs_dir) || {
    error "Cannot resolve remote deploy dir."
    return 1
  }
  fetch_remote_conf || return 1
  install_server "$web_server" "$abs_dir"
}

# ./main.sh webconf [server] — install/reload web config without redeploying.
webconf() {
  select_server "$2" || return 1
  install_webconf || return 1
  patch_conf_toml || return 1
  warn "If conf.toml was patched (app.nginx/media-url), redeploy or restart the backend to apply."
}

# =============================================================================
# Configuration validation & entrypoint
# =============================================================================

# pname becomes a systemd unit name, a tmux session name and a remote dir name.
validate_pname() {
  local n="$1"
  [[ $n =~ ^[a-z0-9][a-z0-9_-]{0,62}$ ]] && return 0
  error "invalid project name '$n' (lowercase letters, digits, '_' or '-', max 63 chars)"
  return 1
}

# Restrict a repo-relative path (app dir / out_dir) to safe characters and
# forbid any '.'/'..' component, so the value stays a path under the repo when
# it is used as a remote path, a zip member and sed replacement text.
# usage: validate_relpath <label> <value>
validate_relpath() {
  local label="$1" val="$2" comp
  [[ $val =~ ^[A-Za-z0-9._/-]+$ ]] || {
    error "$label '$val' may only contain letters, digits, '.', '_', '-' and '/'"
    return 1
  }
  local IFS=/
  for comp in $val; do
    case "$comp" in
    . | ..)
      error "$label '$val' must not contain '.' or '..' components"
      return 1
      ;;
    esac
  done
  return 0
}

# Fail fast on a malformed deploy.conf: every check names the offending entry.
validate_config() {
  validate_pname "$pname" || return 1
  case "$deploy_mode" in tmux | systemctl) ;; *)
    error "deploy_mode must be 'systemctl' or 'tmux' (got '$deploy_mode')"
    return 1
    ;;
  esac
  case "$web_server" in none | nginx | caddy) ;; *)
    error "web_server must be none|nginx|caddy (got '$web_server')"
    return 1
    ;;
  esac
  case "$use_rsync" in true | false) ;; *)
    error "use_rsync must be true|false (got '$use_rsync')"
    return 1
    ;;
  esac

  local entry name path cmd out_dir user host pipes d v a key
  local -A seen=()

  [ ${#servers[@]} -gt 0 ] || {
    error "servers[] is empty"
    return 1
  }
  for entry in "${servers[@]}"; do
    pipes=$(printf '%s' "$entry" | tr -cd '|' | wc -c)
    [ "$pipes" -eq 2 ] || {
      error "servers[] entry '$entry' must be name|ssh_user|host"
      return 1
    }
    IFS='|' read -r name user host <<<"$entry"
    [ -n "$name" ] && [ -n "$user" ] && [ -n "$host" ] || {
      error "servers[] entry '$entry' has an empty field"
      return 1
    }
    case "$host" in *[[:space:]]*)
      error "server '$name' host contains whitespace"
      return 1
      ;;
    esac
    # The target is ssh "$user@$host": a user name with shell metacharacters or
    # a host with a port/colon would change what ssh connects to.
    [[ $user =~ ^[a-z_][a-z0-9_-]*$ ]] || {
      error "server '$name' ssh_user '$user' must match [a-z_][a-z0-9_-]*"
      return 1
    }
    case "$host" in
    -* | *:*)
      error "server '$name' host '$host' must be a hostname or IPv4 address (no port, no IPv6)"
      return 1
      ;;
    esac
    [[ $host =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]] || {
      error "server '$name' host '$host' is not a valid hostname or IPv4 address"
      return 1
      }

    key="s:$name"
    [ -z "${seen[$key]:-}" ] || {
      error "duplicate server '$name' in servers[]"
      return 1
    }
    seen[$key]=1
  done

  [ ${#apps[@]} -gt 0 ] || {
    error "apps[] is empty"
    return 1
  }
  for entry in "${apps[@]}"; do
    pipes=$(printf '%s' "$entry" | tr -cd '|' | wc -c)
    [ "$pipes" -eq 3 ] || {
      error "apps[] entry '$entry' must be name|dir|build_cmd|out_dir"
      return 1
    }
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ -n "$name" ] && [ -n "$path" ] && [ -n "$cmd" ] && [ -n "$out_dir" ] || {
      error "apps[] entry '$entry' has an empty field"
      return 1
    }
    # The name becomes domain_<name>, @<NAME>_DIR@ and assets/<kind>/<name>.conf.
    [[ $name =~ ^[a-z0-9][a-z0-9_-]*$ ]] || {
      error "app name '$name' must match [a-z0-9][a-z0-9_-]*"
      return 1
    }
    case "$name" in api | media)
      error "app name '$name' is reserved (built-in endpoint)"
      return 1
      ;;
    esac
    case "$path" in /*)
      error "app '$name' dir must be relative to the repo root"
      return 1
      ;;
    esac
    case "$out_dir" in /*)
      error "app '$name' out_dir must be relative to its dir"
      return 1
      ;;
    esac
    # Both values end up in remote paths, sed replacement text (compose_webconf)
    # and zip member names, so restrict them and refuse anything that escapes.
    validate_relpath "app '$name' dir" "$path" || return 1
    validate_relpath "app '$name' out_dir" "$out_dir" || return 1
    local real
    real=$(realpath -m -- "$path") || {
      error "app '$name' dir cannot be resolved: $path"
      return 1
    }
    case "$real/" in
    "$PWD/"*) ;;
    *)
      error "app '$name' dir resolves outside the repo: $real"
      return 1
      ;;
    esac
    real=$(realpath -m -- "$path/$out_dir") || {
      error "app '$name' out_dir cannot be resolved: $path/$out_dir"
      return 1
    }
    case "$real/" in
    "$PWD/"*) ;;
    *)
      error "app '$name' out_dir resolves outside the repo: $real"
      return 1
      ;;
    esac
    [ -d "$path" ] || {
      error "app '$name' dir does not exist: $path"
      return 1
    }
    key="a:$name"
    [ -z "${seen[$key]:-}" ] || {
      error "duplicate app '$name' in apps[]"
      return 1
    }
    seen[$key]=1
  done

  # Every domain_* variable is an endpoint: domain_<app> for a site, domain_api
  # and domain_media for the backend/media. Iterating compgen means a new
  # domain_mgr needs no main.sh edit, and a typo is caught here.
  for v in $(compgen -A variable domain_); do
    d="${!v}"
    [ -z "$d" ] && continue
    case "$d" in *://* | */* | *[[:space:]]*)
      error "domain '$d' ($v) must be a bare hostname[:port] (no scheme or path)"
      return 1
      ;;
    esac
    [[ $d =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(:[0-9]+)?$ ]] || {
      error "domain '$d' ($v) is not a valid hostname[:port]"
      return 1
    }
    if [ "$web_server" = "nginx" ]; then
      case "$d" in *:*)
        error "domain '$d' ($v) carries a port; nginx site blocks listen on :80 — drop the port"
        return 1
        ;;
      esac
    fi
    case "$v" in domain_api | domain_media) continue ;; esac
    a="${v#domain_}"
    key="a:$a"
    [ -n "${seen[$key]:-}" ] || {
      error "$v is set but apps[] has no '$a' app"
      return 1
    }
  done

  local k s a
  if [ "${#app_env[@]}" -gt 0 ]; then
    for k in "${!app_env[@]}"; do
      [[ $k =~ ^([^:]+):([^:]+):([A-Z_][A-Z0-9_]*)$ ]] || {
        error "app_env key '$k' must be server:app:VAR"
        return 1
      }
      s="${BASH_REMATCH[1]}"
      a="${BASH_REMATCH[2]}"
      key="s:$s"
      [ -n "${seen[$key]:-}" ] || {
        error "app_env key '$k' references unknown server '$s'"
        return 1
      }
      key="a:$a"
      [ -n "${seen[$key]:-}" ] || {
        error "app_env key '$k' references unknown app '$a'"
        return 1
      }
    done
  fi
  return 0
}

usage() {
  C_RST="\033[0m"
  C_CMD="\033[1;36m"
  C_ARG="\033[33m"
  C_DIM="\033[2m"
  echo "Usage:"
  echo -e "  ${C_CMD}main.sh${C_RST} ${C_ARG}<command>${C_RST} [args]"
  echo ""
  echo "Commands:"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "dev" "Start local dev env (tmux)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "restart" "Restart the tmux session"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "stop" "Stop the tmux session"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "docs" "Initialize/Update Swagger docs"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "lint" "Lint + format gate (Go + scripts)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--fix" "Apply gofumpt/gci formatting first"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "build" "Build backend + SDK + frontends [server]"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "deploy" "Build once, deploy to every server in servers[]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--skip-build" "Deploy without rebuilding (checks .built-for marker)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--force-service" "Overwrite existing service file (systemctl mode)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--allow-placeholder-env" "Deploy even if the API base is still localhost/example.com"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "synconf" "Sync config to server [server]"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "webconf" "Install/reload nginx or caddy site config [server]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "server" "Target server (default: ${servers[0]%%|*})"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "copyfile" "Copy config to local (keeps an existing conf.toml)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--force" "Overwrite conf.toml (wipes jwt secrets)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "servers" "List available servers/apps (see deploy.conf)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "init" "Init project & install dependencies"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "renewal" "Renew project (irreversible)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--no-backup" "Skip the git bundle + conf.toml backup (fully irreversible)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "clean" "Clean project cache"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "serverlog" "Tail remote service log (journalctl -f) [server]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "once" "Print once then exit"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "<lines>" "Number of lines to show (default 100)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "push" "Commit & push to remote"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "initremote" "Create GitHub repo (default: private), bind origin, set upstream (no push)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "[name]" "Repo name (default: $pname)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--public" "Create a public repo (default: private)"
  unset C_RST C_CMD C_ARG C_DIM
}

main() {
  local m="${1:-}"
  case "$m" in
  "" | help | -h | --help)
    usage
    return 0
    ;;
  esac
  validate_config || return 1
  current_cmd="$m"
  trap 'cleanup' EXIT
  trap 'on_interrupt' INT TERM
  case "$m" in
  dev) dev "$@" ;;
  docs) docs "$@" ;;
  lint) lint "${@:2}" ;;
  build) build "$@" ;;
  deploy) deploy "${@:2}" ;;
  synconf) synconf "$@" ;;
  webconf) webconf "$@" ;;
  copyfile) copyfile "$@" ;;
  servers) list "$@" ;;
  init) init "$@" ;;
  renewal) renewal "${@:2}" ;;
  clean) clean "$@" ;;
  serverlog) serverlog "$@" ;;
  push) push "$@" ;;
  initremote) initremote "$@" ;;
  *)
    usage
    return 1
    ;;
  esac
}

# Allow `source main.sh` to load functions only (for validation/debugging) without running a command.
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
