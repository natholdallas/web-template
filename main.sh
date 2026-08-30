#! /usr/bin/env bash

cd "$(dirname "$(readlink -f "$0")")" || exit 1

pname=webtplmst                    # project name
srv_user=xxx                       # server user
srv_host=xxx                       # server host
ssh_target="$srv_user@$srv_host"   # ssh target
tmux_root=tasks                    # tmux root
tmux_dir="$tmux_root/$pname"       # tmux dir
systemd_root=/srv/http             # systemd root dir
systemd_dir="$systemd_root/$pname" # systemd dir
sys=/etc/systemd/system            # systemd service dir
use_rsync="${USE_RSYNC:-true}"     # rsync, scp
tmp_dir="/tmp/$pname"              # tmp dir for deploy
deploy_mode="systemctl"            # deploy target: systemctl, tmux
sudoers_name="webtplmstx"          # shared /etc/sudoers.d filename (merged per-user) — keep fixed across projects

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
  tmux new-window -t "$pname:1" -n "adm" && tmux send-keys -t "$pname:1" 'cd web/apps/adm && pnpm dev' C-m
  tmux new-window -t "$pname:2" -n "usr" && tmux send-keys -t "$pname:2" 'cd web/apps/usr && pnpm dev' C-m
  info "Dev environment started in tmux session: $pname"
}

build() {
  info "Building backend & frontend"
  local p1 p2 p3 rc=0
  go build -o bin/backend &
  p1=$!
  run_in_dir "web" pnpm gen:api || warn "SDK generation failed"
  run_in_dir "web/apps/adm" pnpm generate &
  p2=$!
  run_in_dir "web/apps/usr" pnpm generate &
  p3=$!
  wait "$p1" || rc=1
  wait "$p2" || rc=1
  wait "$p3" || rc=1
  return $rc
}

deploy() {
  local skip_build=false force_service=false
  for arg in "$@"; do
    case "$arg" in
    --skip-build) skip_build=true ;;
    --force-service)
      if [ "$deploy_mode" != "systemctl" ]; then
        warn "--force-service 只在 deploy_mode=systemctl 下生效，已忽略"
      else
        force_service=true
      fi
      ;;
    esac
  done
  case "$deploy_mode" in
  tmux) deploy_tmux "$skip_build" ;;
  systemctl) deploy_systemctl "$skip_build" "$force_service" ;;
  *) error "Unknown deploy_mode '$deploy_mode' (tmux|systemctl)" && return 1 ;;
  esac
}

docs() {
  info "Generating Swagger documentation"
  if ! command -v swag >/dev/null 2>&1; then
    error "swag not found. Install it first: go install github.com/swaggo/swag/cmd/swag@v1.16.6"
    return 1
  fi
  swag fmt || return 1
  swag init --parseDependency --parseInternal || return 1
  info "Regenerating frontend SDK"
  run_in_dir "web" pnpm gen:api
}

# Prepare build artifacts and zip them for deployment.
prepare_deploy() {
  local skip_build="$1"
  if [ "$skip_build" = false ]; then
    build || {
      error "Build failed, aborting deploy."
      return 1
    }
  fi
  for f in bin/backend web/apps/adm/dist web/apps/usr/dist; do
    [ -e "$f" ] || {
      error "Missing build artifact: $f"
      return 1
    }
  done

  # package
  mkdir -p "$tmp_dir"
  rm -f "$tmp_dir/web.zip"
  info "Packaging..."
  zip -r "$tmp_dir/web.zip" bin/backend web/apps/adm/dist web/apps/usr/dist || {
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
  local sudoer_file="$srv_user-$sudoers_name"
  local commands="/usr/bin/systemctl, /bin/mkdir, /bin/chmod, /usr/bin/install, /usr/bin/journalctl"
  local expected="$srv_user ALL=(root) NOPASSWD: $commands"
  # Probe with a whitelisted command that always exits 0 (systemctl is in the list,
  # `sudo -n true` would only succeed for NOPASSWD: ALL users).
  if remote "sudo -n systemctl show-environment >/dev/null 2>&1"; then
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
    remote "sudo mkdir -p $systemd_root && sudo chmod 1777 $systemd_root" || return 1
  fi

  # c. rsync directly into final dir (world-writable, no sudo)
  info "Syncing to $systemd_dir..."
  remote "mkdir -p $systemd_dir"
  remote_util "$tmp_dir/web.zip" "$systemd_dir"
  remote_util "$tmp_dir/$pname.service" "$systemd_dir"
  sync_conf_if_missing "$systemd_dir"

  # d. extract + install service + start + health check (sudo whitelisted)
  info "Extracting & registering systemd service..."
  remote "bash -s" -- "$systemd_dir" "$sys" "$pname" "$force_service" <<'EOF' || return 1
        SRV="$1"; SYS="$2"; P="$3"; FORCE="$4"
        cd "$SRV" || { echo "Directory not found"; exit 1; }
        unzip -qo web.zip && rm -f web.zip && chmod +x bin/backend
        if [ "$FORCE" = "1" ] || [ ! -f "$SYS/${P}.service" ]; then
          sudo install -o root -g root -m 0644 "${SRV}/${P}.service" "$SYS/${P}.service"
          sudo systemctl daemon-reload
        fi
        rm -f "${SRV}/${P}.service"
        sudo systemctl enable "$P"
        sudo systemctl restart "$P"
        for i in $(seq 1 15); do
          sudo systemctl is-active "$P" >/dev/null 2>&1 && exit 0
          sleep 1
        done
        echo "Service $P failed to start" >&2
        exit 1
EOF
  info "Systemctl deployment successful."
}

synconf() {
  remote "mkdir -p $tmux_dir"
  remote_util conf.toml "$tmux_dir"
}

copyfile() {
  cp ./assets/conf.toml ./conf.toml
  cp ./assets/nuxt.env web/apps/adm/.env
  cp ./assets/nuxt.env web/apps/adm/.env.production
  cp ./assets/nuxt.env web/apps/usr/.env
  cp ./assets/nuxt.env web/apps/usr/.env.production
}

init() {
  go install github.com/swaggo/swag/cmd/swag@v1.16.6
  go install github.com/silenceper/gowatch@latest
  go install github.com/gofiber/cli/fiber@latest
  git submodule update --init --recursive
  go mod tidy
  run_in_dir "web" pnpm install
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
    # copy file
    copyfile
    # reset git directory
    rm -rf .git
    rm -rf web/packages/natholdallas
    rm .gitmodules
    # replace all occurrences of old_name with new_name (escape sed special chars)
    if [[ "$OSTYPE" == "darwin"* ]]; then
      # macOS 语法
      SED_CMD=(sed -i "")
    else
      # Linux 语法
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
      -o -name "assets" \
      -o -name "logs" \
      -o -name "media" \
      -o -name "bin" \
      -o -name "go.sum" \
      -o -name "pnpm-lock.yaml" \) -prune \
      -o -type f \
      \( -name "*.go" -o -name "*.md" -o -name "*.toml" \
      -o -name "*.service" -o -name "*.sh" -o -name "*.json" \
      -o -name "*.yml" -o -name "*.yaml" -o -name "*.ts" \
      -o -name "*.tsx" -o -name "*.vue" -o -name "*.nuxtrc" \
      -o -name "*.env" -o -name ".gitmodules" \) \
      -exec "${SED_CMD[@]}" "s/${pname_esc}/${new_name_esc}/g" {} +
    # initialize git repository
    git init
    git submodule add https://github.com/natholdallas/nuxt-modules.git web/packages/natholdallas
    git add -A
    # generate docs
    docs
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
  local lines="${2:-100}"
  if [ "$2" = "once" ]; then
    ssh -t "$ssh_target" "sudo journalctl -u $pname -n $lines"
    return
  fi
  ssh -t "$ssh_target" "sudo journalctl -u $pname -n $lines -f"
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
  if [ -z "$branch" ] || ! git branch -r | grep -q "origin/$branch$"; then
    warn "No upstream for branch '$branch'. Set upstream: git push -u origin $branch"
  fi
  git push
}

mode=$1
case "$1" in
dev) dev "$@" ;;
docs) docs "$@" ;;
build) build "$@" ;;
deploy) deploy "${@:2}" ;;
synconf) synconf "$@" ;;
copyfile) copyfile "$@" ;;
init) init "$@" ;;
renewal) renewal "$@" ;;
clean) clean "$@" ;;
serverlog) serverlog "$@" ;;
push) push "$@" ;;
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
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "build" "Compile Go backend + generate static sites"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "deploy" "Build, sync, deploy per deploy_mode (tmux|systemctl)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--skip-build" "Deploy without rebuilding"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "--force-service" "Overwrite existing service file (systemctl mode)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "synconf" "Sync config to server"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "copyfile" "Copy config to local"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "init" "Init project & install dependencies"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "renewal" "Renew project (irreversible)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "clean" "Clean project cache"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "serverlog" "Tail remote service log (journalctl -f)"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "once" "Print once then exit"
  printf "  ${C_DIM}  %-16s${C_RST}         %s\n" "<lines>" "Number of lines to show (default 100)"
  printf "  ${C_CMD}%-18s${C_RST}     %s\n" "push" "Commit & push to remote"
  unset C_RST C_CMD C_ARG C_DIM
  ;;
esac
