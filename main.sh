#! /usr/bin/env bash

pname=webtplmst
host=xxx@xxx
dir=tasks/$pname
use_rynsc=true

run_in_dir() { (cd "$1" && shift && "$@"); }
success() { echo -e "\033[0;32m[SUCCESS]\033[0m $1"; }
info() { echo -e "\033[1;36m[INFO]\033[0m $1"; }
warn() { echo -e "\033[1;33m[WARN]\033[0m $1"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1"; }
remote_util() {
  if $use_rynsc; then
    info "use rsync to sync files..."
    rsync -avP "$1" $host:$dir
  else
    info "use scp to sync files..."
    scp "$1" $host:$dir
  fi
}

dev() {
  if [ "$2" = "--force" ]; then
    warn "Forcing restart of tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
  fi

  if tmux has-session -t "$pname" 2>/dev/null; then
    warn "Session $pname already exists."
    return
  fi

  tmux new-session -d -s "$pname" -n "go"
  tmux new-window -t "$pname:1" -n "adm"
  tmux new-window -t "$pname:2" -n "usr"

  sleep 0.2

  tmux send-keys -t "$pname:0" 'gowatch -o bin/backend' C-m
  tmux send-keys -t "$pname:1" 'cd web/apps/adm && pnpm dev' C-m
  tmux send-keys -t "$pname:2" 'cd web/apps/usr && pnpm dev' C-m

  info "Dev environment started in tmux session: $pname"
}

build() {
  info "Building backend & frontend"
  go build -o bin/backend &
  run_in_dir "web/apps/adm" pnpm generate &
  run_in_dir "web/apps/usr" pnpm generate &
  wait
}

docs() {
  info "Generating Swagger documentation"
  swag fmt
  swag init --parseDependency --parseInternal
}

deploy() {
  if [ "$2" != "--skip-build" ]; then
    build
  fi

  # package
  info "Packaging..."
  zip -r web.zip bin/backend web/apps/adm/dist web/apps/usr/dist

  # remote
  ssh $host mkdir -p $dir
  remote_util web.zip
  rm web.zip

  # remote shell
  info "Restarting remote service..."
  ssh "$host" "bash -s" -- "$dir" "$pname" <<'EOF'
        REMOTE_DIR="$1"
        REMOTE_PNAME="$2"
        cd "$REMOTE_DIR" || { echo "Directory not found"; exit 1; }
        tmux kill-session -t "$REMOTE_PNAME" 2>/dev/null
        tmux new-session -d -s "$REMOTE_PNAME" -n "server"
        tmux send-keys -t "$REMOTE_PNAME" "unzip -qo web.zip && rm web.zip" C-m
        tmux send-keys -t "$REMOTE_PNAME" "chmod +x ./bin/backend && ./bin/backend" C-m
EOF
  info "Deployment successful."
}

synconf() {
  ssh $host mkdir -p $dir
  remote_util conf.toml
}

init() {
  if [ "$2" = "--copy-file" ]; then
    cp ./assets/conf.toml ./conf.toml
    cp ./assets/nuxt.env web/apps/adm/.env
    cp ./assets/nuxt.env web/apps/adm/.env.production
    cp ./assets/nuxt.env web/apps/usr/.env
    cp ./assets/nuxt.env web/apps/usr/.env.production
  fi
  go install github.com/swaggo/swag/cmd/swag@latest
  go install github.com/silenceper/gowatch@latest
  go install github.com/gofiber/cli/fiber@latest
  git submodule update --init --recursive
  go mod tidy
  run_in_dir "web" pnpm install
}

renewal() {
  local old_name="$pname"
  local new_name="$2"

  if [ -z "$new_name" ]; then
    warn "Missing second parameter."
    exit 0
  fi

  echo -e "Project Initialization & Reset"
  echo "--------------------------------------------------"
  echo "This script will perform the following IRREVERSIBLE actions:"
  echo -e "1. GLOBAL REPLACE: All occurrences of '$old_name' will be changed to '$new_name'."
  echo -e "2. GIT RESET: The existing .git directory will be DELETED."
  echo -e "3. RE-INIT: A new git repository will be initialized in this directory."
  echo -e "4. PROTECTION: 'main.sh' will be skipped during string replacement."
  echo "--------------------------------------------------"

  # prompt for user confirmation
  read -rp "Are you sure you want to proceed? (y/N): " confirm
  if [[ "$confirm" =~ ^[yY](es)?$ ]]; then
    # reset git directory
    rm -rf .git
    rm -rf web/packages/natholdallas
    rm .gitmodules
    # replace all occurrences of old_name with new_name
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
      -o -name "bin" \) -prune \
      -o -type f \
      -exec sed -i "s/${old_name}/${new_name}/g" {} +
    # generate docs
    docs
    # initialize git repository
    git init
    git submodule add https://github.com/natholdallas/nuxt-modules.git web/packages/natholdallas
    git add -A
    success "Project initialized successfully."
  fi
}

clean() {
  find . -name "node_modules" -type d -prune -exec rm -rf {} +
  find . -name "dist" -type d -prune -exec rm -rf {} +
  find . -name ".nuxt" -type d -prune -exec rm -rf {} +
  find . -name ".output" -type d -prune -exec rm -rf {} +
}

case "$1" in
dev) dev "$@" ;;
docs) docs "$@" ;;
build) build "$@" ;;
deploy) deploy "$@" ;;
synconf) synconf "$@" ;;
init) init "$@" ;;
renewal) renewal "$@" ;;
clean) clean "$@" ;;
*)
  echo "Usage:"
  echo "  dev:          Start local development environment (tmux) "
  echo "  docs:         Initialize/Update Swagger documentation "
  echo "  build:        Compile Go backend and generate static sites "
  echo "  deploy:       Build, sync to server, and hot-reload via tmux "
  echo "  synconf:      Sync config to server "
  echo "  init:         Init Project & Install dependencies "
  echo "  renewal:      Renewal project "
  echo "  clean:        Clean Project Cache "
  exit 1
  ;;
esac
