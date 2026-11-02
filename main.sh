#! /usr/bin/env bash

cd "$(dirname "$(readlink -f "$0")")" || exit 1

# ============================== CONFIGURATION (edit only here) ==============================
pname=webtplmst                     # project name (systemd unit / tmux session / remote dir name)
deploy_mode="systemctl"             # deploy target: systemctl | tmux
use_rsync="${USE_RSYNC:-true}"      # transfer method: true=rsync, false=scp
web_server="none"                   # web server: none | nginx | caddy
web_api_addr="127.0.0.1:8080"       # api reverse-proxy target (must match conf.toml app.port)
client_max_body="200m"              # reverse-proxy body limit (must match conf.toml app.body-limit)

# remote deploy dirs
tmux_root=tasks                     # tmux-mode root dir (relative to $HOME)
systemd_root=/srv/http              # systemd-mode root dir
sys=/etc/systemd/system              # systemd service dir
sudoers_name="webtplmstx"           # /etc/sudoers.d file name (shared & fixed — do not rename)

# registries — entry formats & examples: README "main.sh configuration guide"
apps=( # format: name|dir|build_cmd|out_dir
  "adm|web/apps/adm|pnpm generate|dist"
  "usr|web/apps/usr|pnpm generate|dist"
)
servers=( # format: name|ssh_user|host (default server = first entry)
  "prod|deploy-user|203.0.113.72"
)
declare -A app_env                  # format: server:app:VAR (explicit env overrides win)
domains=( # format: endpoint|domain (usr|adm|api|media; empty = endpoint disabled)
  "usr|"
  "adm|"
  "api|"
  "media|"
)
app_bind=( # format: app|api_endpoint|site_endpoint (drives env auto-derivation)
  "usr|api|usr"
  "adm|api|adm"
)

# ===================== DERIVED / RUNTIME STATE (computed, do not edit) =====================
tmux_dir="$tmux_root/$pname"        # tmux deploy dir
systemd_dir="$systemd_root/$pname"  # systemd deploy dir
tmp_dir="/tmp/$pname"               # packaging temp dir
srv_name=""                         # selected server (set by select_server)
srv_user=""
srv_host=""
ssh_target=""                        # ssh target
SUDO=""                              # remote command prefix (root skips sudo)

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
  if $use_rsync; then
    info "use rsync to sync files..."
    rsync -avP "$src" "$ssh_target:$target"
  else
    info "use scp to sync files..."
    scp "$src" "$ssh_target:$target"
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
  info "Web server: $web_server"
  if [ "$web_server" != "none" ]; then
    info "Domain bindings:"
    for entry in "${domains[@]}"; do
      IFS='|' read -r name value <<<"$entry"
      [ -n "$value" ] && echo "  $name  ->  $value"
    done
  fi
}

dev() {
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
    tmux new-window -t "$pname:$i" -n "$name" && tmux send-keys -t "$pname:$i" "cd $path && pnpm dev" C-m
    i=$((i + 1))
  done
  info "Dev environment started in tmux session: $pname"
}

# Build backend + SDK (server-independent). Run once per deploy.
build_core() {
  info "Building backend & SDK"
  local p1 rc=0
  go build -o bin/backend &
  p1=$!
  run_in_dir "web" pnpm gen:api || warn "SDK generation failed"
  wait "$p1" || rc=1
  return $rc
}

# Build every app in parallel, exporting the selected server's env overrides.
build_front() {
  info "Building frontends for $srv_name ..."
  local entry name path cmd out_dir k cmd_args=() envs=() pids=() rc=0
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ -d "$path" ] || {
      error "App directory not found: $path"
      return 1
    }
    read -ra cmd_args <<<"$cmd"
    envs=()
    for k in "${!app_env[@]}"; do
      [[ "$k" == "$srv_name:$name:"* ]] && envs+=("${k#"$srv_name:$name:"}=${app_env[$k]}")
    done
    auto_envs "$name" envs
    info "  building $name ($path)"
    (cd "$path" && env "${envs[@]}" "${cmd_args[@]}") &
    pids+=($!)
  done
  for p in "${pids[@]}"; do wait "$p" || rc=1; done
  if [ $rc -eq 0 ]; then
    for entry in "${apps[@]}"; do
      IFS='|' read -r name path cmd out_dir <<<"$entry"
      echo "$srv_name" >"$path/$out_dir/.built-for"
    done
  fi
  return $rc
}

# Backend + SDK + frontends for the selected server, in parallel.
build() {
  local p rc=0
  select_server || return 1
  build_core &
  p=$!
  build_front || rc=1
  wait "$p" || rc=1
  return $rc
}

deploy() {
  local skip_build=false force_service=false arg
  for arg in "$@"; do
    case "$arg" in
    --skip-build) skip_build=true ;;
    --force-service)
      if [ "$deploy_mode" != "systemctl" ]; then
        warn "--force-service only applies in deploy_mode=systemctl, ignored"
      else
        force_service=true
      fi
      ;;
    -*)
      warn "Unknown option: $arg (ignored)"
      ;;
    *)
      warn "Unexpected argument: $arg (ignored) — deploy ships to every server in servers[]"
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
    case "$deploy_mode" in
    tmux) deploy_tmux "$skip_build" || rc=1 ;;
    systemctl) deploy_systemctl "$skip_build" "$force_service" || rc=1 ;;
    *) error "Unknown deploy_mode '$deploy_mode' (tmux|systemctl)" && return 1 ;;
    esac
    install_webconf || rc=1
  done
  [ $rc -eq 0 ] && success "Deployed to all servers."
  return $rc
}

# Warn when --skip-build reuses artifacts built for a different server.
check_build_marker() {
  local entry name path cmd out_dir built
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    if [ -f "$path/$out_dir/.built-for" ]; then
      built=$(<"$path/$out_dir/.built-for")
      if [ "$built" != "$srv_name" ]; then
        warn "$name artifacts were built for '$built', target is '$srv_name', --skip-build will deploy the wrong artifacts!"
      fi
    else
      warn "$name artifacts lack a build marker (maybe never built via deploy)"
    fi
  done
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
  local -a files=(bin/backend)
  local entry name path cmd out_dir
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ -e "$path/$out_dir" ] || {
      error "Missing build artifact: $path/$out_dir"
      return 1
    }
    files+=("$path/$out_dir")
  done

  # package
  mkdir -p "$tmp_dir"
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
  remote "mkdir -p $tmux_dir"
  sync_conf_if_missing "$tmux_dir"
  patch_conf_toml
  remote_util "$tmp_dir/web.zip" "$tmux_dir"
  rm "$tmp_dir/web.zip"

  # remote shell
  info "Restarting remote service..."
  remote "bash -s" -- "$tmux_dir" "$pname" <<'EOF' || return 1
        REMOTE_DIR="$1"
        REMOTE_PNAME="$2"
        cd "$REMOTE_DIR" || { echo "Directory not found"; exit 1; }
        tmux kill-session -t "$REMOTE_PNAME" 2>/dev/null
        tmux new-session -d -s "$REMOTE_PNAME" -n "server"
        tmux new-window -t "$REMOTE_PNAME:1" -n "panel"
        tmux send-keys -t "$REMOTE_PNAME:0" "unzip -qo web.zip && rm web.zip && chmod +x ./bin/backend && ./bin/backend" C-m
EOF
  info "Tmux deployment successful."
}

sync_conf_if_missing() {
  local remote_dir="$1"
  if ! remote "test -f $remote_dir/conf.toml"; then
    info "No remote conf.toml, syncing local conf.toml..."
    remote_util conf.toml "$remote_dir"
  else
    info "Remote conf.toml exists, skipping sync to avoid overwrite."
  fi
}

# Ensure remote has limited passwordless sudo for the commands this deploy needs.
# All projects on the same server share one file: /etc/sudoers.d/<user>-<sudoers_name>.
# Idempotent: skips if already configured. Validates the sudoers file first.
ensure_sudoers() {
  [ -n "$SUDO" ] || {
    info "srv_user=root, skipping sudoers setup."
    return 0
  }
  local sudoer_file="$srv_user-$sudoers_name"
  local commands="/usr/bin/systemctl, /bin/mkdir, /bin/chmod, /usr/bin/install, /usr/bin/journalctl, /usr/sbin/nginx, /usr/bin/caddy, /usr/bin/tee"
  local expected="$srv_user ALL=(root) NOPASSWD: $commands"
  # Probe the whitelisted commands actually usable (systemctl is the base marker;
  # when web_server != none also verify nginx/caddy/tee so existing servers get
  # the extended sudoers file re-installed instead of skipping).
  local probe="sudo -n systemctl show-environment >/dev/null 2>&1"
  [ "$web_server" = "nginx" ] && probe="$probe && sudo -n nginx -v >/dev/null 2>&1"
  [ "$web_server" = "caddy" ] && probe="$probe && sudo -n caddy version >/dev/null 2>&1 && echo x | sudo -n tee /dev/null >/dev/null 2>&1"
  if remote "$probe"; then
    info "sudoers $sudoer_file already configured, skipping."
    return 0
  fi
  warn "Configuring limited passwordless sudo for $srv_user ($commands)."
  echo "$expected" | remote "cat > /tmp/$sudoer_file"
  ssh -t "$ssh_target" "sudo visudo -cf /tmp/$sudoer_file && sudo install -o root -g root -m 0440 /tmp/$sudoer_file /etc/sudoers.d/$sudoer_file && rm -f /tmp/$sudoer_file"
  info "sudoers configured."
}

# Full systemd deploy: sync build into /srv/http, install service, start.
# Requires limited passwordless sudo for systemctl/mkdir/chmod/cp.
deploy_systemctl() {
  local skip_build="$1" force_service="$2"
  prepare_deploy "$skip_build" || return 1
  ensure_sudoers

  # a. render service file into tmp_dir
  sed -e "s/@PNAME@/$pname/g" -e "s|@SRV_DIR@|$systemd_dir|g" assets/run.service >"$tmp_dir/$pname.service"

  # b. ensure $systemd_root exists & is sticky/world-writable (only on first setup)
  if ! remote "stat -c %a $systemd_root 2>/dev/null | grep -qx 1777"; then
    info "First-time setup: creating $systemd_root with mode 1777..."
    remote "$SUDO mkdir -p $systemd_root && $SUDO chmod 1777 $systemd_root" || return 1
  fi

  # c. rsync directly into final dir (world-writable, no sudo)
  info "Syncing to $systemd_dir..."
  remote "mkdir -p $systemd_dir"
  remote_util "$tmp_dir/web.zip" "$systemd_dir"
  remote_util "$tmp_dir/$pname.service" "$systemd_dir"
  sync_conf_if_missing "$systemd_dir"
  patch_conf_toml

  # d. extract + install service + start + health check (sudo whitelisted)
  info "Extracting & registering systemd service..."
  remote "bash -s" -- "$systemd_dir" "$sys" "$pname" "$force_service" "$SUDO" <<'EOF' || return 1
        SRV="$1"; SYS="$2"; P="$3"; FORCE="$4"; SUDO="$5"
        cd "$SRV" || { echo "Directory not found"; exit 1; }
        unzip -qo web.zip && rm -f web.zip && chmod +x bin/backend
        if [ "$FORCE" = "1" ] || [ ! -f "$SYS/${P}.service" ]; then
          $SUDO install -o root -g root -m 0644 "${SRV}/${P}.service" "$SYS/${P}.service"
          $SUDO systemctl daemon-reload
        fi
        rm -f "${SRV}/${P}.service"
        $SUDO systemctl enable "$P"
        $SUDO systemctl restart "$P"
        for i in $(seq 1 15); do
          $SUDO systemctl is-active "$P" >/dev/null 2>&1 && exit 0
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
  remote "mkdir -p $d"
  remote_util conf.toml "$d"
}

copyfile() {
  if [ -f ./assets/conf.toml ]; then
    cp ./assets/conf.toml ./conf.toml
  else
    info "No assets/conf.toml template, skipping (create conf.toml manually)."
  fi
  local entry name path cmd out_dir
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    cp ./assets/nuxt.env "$path/.env"
    cp ./assets/nuxt.env "$path/.env.production"
  done
}

init() {
  command -v gowatch >/dev/null 2>&1 || go install github.com/silenceper/gowatch@latest
  if ! command -v swag >/dev/null 2>&1; then
    go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5
    return 1
  fi
  git submodule update --init --recursive
  docs
  go mod tidy
  run_in_dir "web" pnpm install
}

docs() {
  info "Generating Swagger documentation"
  swag fmt || return 1
  swag init --parseDependency --parseInternal || return 1
  info "Regenerating frontend SDK"
  run_in_dir "web" pnpm gen:api
}

renewal() {
  local new_name=$2 SED_CMD=()
  if [ -z "$2" ]; then
    read -rp "Input new project name: " new_name
  fi
  if [ -z "$new_name" ]; then
    warn "Missing Project Name."
    return 0
  fi

  echo -e "Project Initialization & Reset"
  echo "--------------------------------------------------"
  echo "This script will perform the following IRREVERSIBLE actions:"
  echo -e "1. GLOBAL REPLACE: All occurrences of '$pname' (including main.sh) will be changed to '$new_name'."
  echo -e "2. GIT RESET: The existing .git directory will be DELETED."
  echo -e "3. RE-INIT: A new git repository will be initialized in this directory."
  echo -e "4. CONFIG: conf.toml and .env files will be recreated from assets/."
  echo "--------------------------------------------------"

  # prompt for user confirmation
  read -rp "Are you sure you want to proceed? (y/N): " confirm
  if [[ "$confirm" =~ ^[yY](es)?$ ]]; then
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
      -o -name "pnpm-lock.yaml" \) -prune \
      -o -type f \
      \( -name "*.go" -o -name "*.md" -o -name "*.mod" -o -name "*.toml" \
      -o -name "*.service" -o -name "*.sh" -o -name "*.json" \
      -o -name "*.yml" -o -name "*.yaml" -o -name "*.ts" \
      -o -name "*.tsx" -o -name "*.vue" -o -name "*.nuxtrc" \
      -o -name "*.env" -o -name ".gitmodules" \) \
      -exec "${SED_CMD[@]}" "s/${pname_esc}/${new_name_esc}/g" {} +
    # initialize git repository
    git init
    git submodule add https://github.com/natholdallas/nuxt-modules.git web/packages/natholdallas
    # copy file
    copyfile
    # initialize project
    init
    # add all files & commit
    git add -A
    git commit -m "initialize"
    success "Project initialized successfully."
  fi
}

cleanup() {
  rm -f "$tmp_dir/web.zip" "$tmp_dir/$pname.service"
  [ -d "$tmp_dir" ] && rm -rf "$tmp_dir" && info "Cleaned up $tmp_dir"
}

on_interrupt() {
  if [ "$mode" = "dev" ] && tmux has-session -t "$pname" 2>/dev/null; then
    warn "Stopping tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
  fi
  exit 1
}

trap 'cleanup' EXIT
trap 'on_interrupt' INT TERM

serverlog() {
  local server="" arg="$2"
  if [ -n "$arg" ] && is_server "$arg"; then
    server="$arg"
    arg="$3"
  fi
  select_server "$server" || return 1
  local lines="${arg:-100}"
  if [ "$arg" = "once" ]; then
    ssh -t "$ssh_target" "$SUDO journalctl -u $pname -n $lines"
  else
    ssh -t "$ssh_target" "$SUDO journalctl -u $pname -n $lines -f"
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
  for arg in "${@:2}"; do
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

# Print the domain for an endpoint in domains[] (empty when unset).
domain_of() {
  local ep="$1" entry key val
  for entry in "${domains[@]}"; do
    IFS='|' read -r key val <<<"$entry"
    [ "$key" = "$ep" ] && {
      printf '%s' "$val"
      return 0
    }
  done
  return 1
}

# Print configured domains as "endpoint=domain" list (for success messages).
get_domain_list() {
  local entry ep d out=()
  for entry in "${domains[@]}"; do
    IFS='|' read -r ep d <<<"$entry"
    [ -n "$d" ] && out+=("$ep=$d")
  done
  printf '%s' "${out[*]}"
}

# Append auto-derived env vars for an app (NUXT_PUBLIC_API_BASE / SITE_URL)
# from domains[] + app_bind[]. Explicit entries already in the envs array win.
# usage: auto_envs <app> <envs_array_name>
auto_envs() {
  local app="$1"
  local -n _envs="$2"
  local bind app_ep api_ep site_ep api_d site_d v found
  for bind in "${app_bind[@]}"; do
    IFS='|' read -r app_ep api_ep site_ep <<<"$bind"
    [ "$app_ep" = "$app" ] || continue
    api_d=$(domain_of "$api_ep")
    site_d=$(domain_of "$site_ep")
    if [ -n "$api_d" ]; then
      found=0
      for v in "${_envs[@]}"; do [[ "${v%%=*}" = "NUXT_PUBLIC_API_BASE" ]] && found=1; done
      [ $found -eq 0 ] && _envs+=("NUXT_PUBLIC_API_BASE=https://$api_d")
    fi
    if [ -n "$site_d" ]; then
      found=0
      for v in "${_envs[@]}"; do [[ "${v%%=*}" = "NUXT_PUBLIC_SITE_URL" ]] && found=1; done
      [ $found -eq 0 ] && _envs+=("NUXT_PUBLIC_SITE_URL=https://$site_d")
    fi
    return
  done
}

# Resolve the absolute deploy dir on the remote (nginx/caddy root needs absolutes).
remote_abs_dir() {
  if [ "$deploy_mode" = "tmux" ]; then
    remote "cd $tmux_dir && pwd" || return 1
  else
    deploy_dir
  fi
}

# Compose the web server config into $tmp_dir/$pname.<fext> from assets/<kind>/*,
# substituting domain + deploy-dir tokens. Skips endpoints whose domain is empty.
# usage: compose_webconf <nginx|caddy> <abs_deploy_dir>   (returns 2 if all empty)
compose_webconf() {
  local kind="$1" abs_dir="$2"
  local fext=conf
  [ "$kind" = "caddy" ] && fext=caddy
  local name path cmd out_dir entry ep domain usr_dir="" adm_dir=""
  # shellcheck disable=SC2178
  local media_dir="$abs_dir/media" fname out=""
  for entry in "${apps[@]}"; do
    IFS='|' read -r name path cmd out_dir <<<"$entry"
    [ "$name" = "usr" ] && usr_dir="$abs_dir/$path/$out_dir/public"
    [ "$name" = "adm" ] && adm_dir="$abs_dir/$path/$out_dir/public"
  done
  if [ -n "$(domain_of usr)" ] && [ -z "$usr_dir" ]; then
    error "usr domain set but no 'usr' app in apps[]"
    return 1
  fi
  if [ -n "$(domain_of adm)" ] && [ -z "$adm_dir" ]; then
    error "adm domain set but no 'adm' app in apps[]"
    return 1
  fi
  for entry in "${domains[@]}"; do
    IFS='|' read -r ep domain <<<"$entry"
    [ -n "$domain" ] || continue
    fname="assets/$kind/$ep.$fext"
    [ -f "$fname" ] || {
      error "Missing web template: $fname"
      return 1
    }
    # shellcheck disable=SC2179
    out+=$(sed -e "s|@DOMAIN@|$domain|g" -e "s|@SRV_DIR@|$abs_dir|g" \
      -e "s|@USR_DIR@|$usr_dir|g" -e "s|@ADM_DIR@|$adm_dir|g" \
      -e "s|@MEDIA_DIR@|$media_dir|g" -e "s|@API_ADDR@|$web_api_addr|g" \
      -e "s|@CLIENT_MAX_BODY@|$client_max_body|g" "$fname")$'\n'
  done
  # shellcheck disable=SC2128
  [ -n "$out" ] || {
    warn "domains[] all empty, skipping web config generation."
    return 2
  }
  printf '%s' "$out" >"$tmp_dir/$pname.$fext"
}

# Idempotently patch the remote conf.toml for web-server fronting:
#   app.nginx = true            (backend stops serving static)
#   resources.media-url = "..." (only when the media domain is set)
patch_conf_toml() {
  [ "$web_server" = "none" ] && return 0
  local remote_conf media_d
  remote_conf="$(deploy_dir)/conf.toml"
  media_d=$(domain_of media)
  remote "test -f $remote_conf" || {
    warn "No remote conf.toml at $remote_conf, skipping patch."
    return 0
  }
  remote "bash -s" -- "$remote_conf" "$media_d" <<'EOF'
    CONF="$1"; MEDIA_D="$2"
    changed=0
    if ! grep -q '^nginx[[:space:]]*=[[:space:]]*true' "$CONF"; then
      sed -i -E 's/^nginx([[:space:]]*)=([[:space:]]*)false/nginx\1=\2true/' "$CONF" && changed=1
    fi
    if [ -n "$MEDIA_D" ] && ! grep -q "media-url[[:space:]]*=.*$MEDIA_D" "$CONF"; then
      sed -i -E "s|^media-url[[:space:]]*=.*|media-url = \"https://$MEDIA_D\" # auto (main.sh)|" "$CONF" && changed=1
    fi
    [ "$changed" = "1" ] && echo "conf.toml patched (app.nginx / media-url)"
EOF
}

# Install & reload the web server site config for the given kind (nginx|caddy).
install_server() {
  local kind="$1" abs_dir="$2" bin fext
  case "$kind" in
  nginx) bin=nginx fext=conf ;;
  caddy) bin=caddy fext=caddy ;;
  *) error "Unknown web server kind '$kind' (nginx|caddy)" && return 1 ;;
  esac
  remote "command -v $bin >/dev/null 2>&1" || {
    error "$bin not installed on $ssh_target — run: sudo apt install $bin"
    return 1
  }
  compose_webconf "$kind" "$abs_dir" || return 0 # 2 = no domains, skip
  remote "mkdir -p $tmp_dir"
  remote_util "$tmp_dir/$pname.$fext" "$tmp_dir"
  remote "bash -s" -- "$tmp_dir" "$pname" "$SUDO" "$bin" "$fext" <<'EOF' || return 1
    T="$1"; P="$2"; SUDO="$3"; BIN="$4"; EXT="$5"
    case "$BIN" in
    nginx)
      $SUDO install -o root -g root -m 0644 "$T/${P}.${EXT}" "/etc/nginx/conf.d/${P}.${EXT}"
      if $SUDO nginx -t >/dev/null 2>&1; then
        $SUDO systemctl reload nginx
      else
        $SUDO nginx -t
        echo "nginx config test FAILED, ${P}.${EXT} kept at /etc/nginx/conf.d/${P}.${EXT}" >&2
        exit 1
      fi
      ;;
    caddy)
      $SUDO mkdir -p /etc/caddy/conf.d
      $SUDO install -o root -g root -m 0644 "$T/${P}.${EXT}" "/etc/caddy/conf.d/${P}.${EXT}"
      if ! grep -qs 'import conf.d/\*\.caddy' /etc/caddy/Caddyfile; then
        printf '\nimport conf.d/*.caddy\n' | $SUDO tee -a /etc/caddy/Caddyfile >/dev/null
      fi
      if $SUDO caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null 2>&1; then
        $SUDO systemctl reload caddy
      else
        $SUDO caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
        echo "caddy validate FAILED, ${P}.${EXT} kept at /etc/caddy/conf.d/${P}.${EXT}" >&2
        exit 1
      fi
      ;;
    esac
    rm -f "$T/${P}.${EXT}"
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
  abs_dir=$(remote_abs_dir) || {
    error "Cannot resolve remote deploy dir."
    return 1
  }
  install_server "$web_server" "$abs_dir"
}

# ./main.sh webconf [server] — install/reload web config without redeploying.
webconf() {
  select_server "$2" || return 1
  patch_conf_toml
  install_webconf || return 1
  warn "If conf.toml was patched (app.nginx/media-url), redeploy or restart the backend to apply."
}

mode=$1
case "$1" in
dev) dev "$@" ;;
docs) docs "$@" ;;
build) build "$@" ;;
deploy) deploy "${@:2}" ;;
synconf) synconf "$@" ;;
webconf) webconf "$@" ;;
copyfile) copyfile "$@" ;;
servers) list "$@" ;;
init) init "$@" ;;
renewal) renewal "$@" ;;
clean) clean "$@" ;;
serverlog) serverlog "$@" ;;
push) push "$@" ;;
initremote) initremote "$@" ;;
*)
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
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "build" "Build backend + SDK + frontends (default server)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "deploy" "Build once, deploy to every server in servers[]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--skip-build" "Deploy without rebuilding (checks .built-for marker)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--force-service" "Overwrite existing service file (systemctl mode)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "synconf" "Sync config to server [server]"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "webconf" "Install/reload nginx or caddy site config [server]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "server" "Target server (default: ${servers[0]%%|*})"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "copyfile" "Copy config to local"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "servers" "List available servers/apps (see registries)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "init" "Init project & install dependencies"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "renewal" "Renew project (irreversible)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "clean" "Clean project cache"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "serverlog" "Tail remote service log (journalctl -f) [server]"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "once" "Print once then exit"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "<lines>" "Number of lines to show (default 100)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "push" "Commit & push to remote"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "initremote" "Create GitHub repo (default: private), bind origin, set upstream (no push)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "[name]" "Repo name (default: $pname)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--public" "Create a public repo (default: private)"
  unset C_RST C_CMD C_ARG C_DIM
  ;;
esac
