#! /usr/bin/env bash

pname=webtplmst
host=xxx@xxx
dir=tasks/$pname

run_in_dir() {
  (cd "$1" && shift && "$@")
}

dev() {
  if [ "$2" = "--force" ]; then
    echo "Forcing restart of tmux session: $pname"
    tmux kill-session -t "$pname" 2>/dev/null
  fi

  if tmux has-session -t "$pname" 2>/dev/null; then
    echo "Session $pname already exists."
    return
  fi

  tmux new-session -d -s "$pname" -n "backend"
  tmux new-window -t "$pname:1" -n "admin"
  tmux new-window -t "$pname:2" -n "user"

  sleep 0.2

  tmux send-keys -t "$pname:0" 'gowatch -o bin/backend' C-m
  tmux send-keys -t "$pname:1" 'cd web/apps/admin && pnpm dev' C-m
  tmux send-keys -t "$pname:2" 'cd web/apps/user && pnpm dev' C-m

  echo "Dev environment started in tmux session: $pname"
}

build() {
  echo "Building backend..."
  go build -o bin/backend

  echo "Generating frontend static files..."
  run_in_dir "web/apps/admin" pnpm generate
  run_in_dir "web/apps/user" pnpm generate
}

docs() {
  swag fmt
  swag init --parseDependency --parseInternal
}

deploy() {
  if [ "$2" != "--skip-build" ]; then
    build
  fi

  # prepare
  ssh $host mkdir -p $dir

  # package
  echo "Packaging..."
  zip -r web.zip bin/backend web/apps/admin/dist

  # remote
  rsync -avP web.zip $host:$dir
  rm web.zip

  # remote shell
  echo "Restarting remote service..."
  ssh "$host" "bash -s" -- "$dir" "$pname" <<'EOF'
        # 在远程环境获取参数
        REMOTE_DIR="$1"
        REMOTE_PNAME="$2"

        cd "$REMOTE_DIR" || { echo "Directory not found"; exit 1; }

        # 清理旧会话并开启新会话
        tmux kill-session -t "$REMOTE_PNAME" 2>/dev/null
        tmux new-session -d -s "$REMOTE_PNAME" -n "server"
        
        # 发送解压与启动命令
        tmux send-keys -t "$REMOTE_PNAME" "unzip -qo web.zip && rm web.zip" C-m
        tmux send-keys -t "$REMOTE_PNAME" "chmod +x ./bin/backend && ./bin/backend" C-m
EOF
  echo "Deployment successful."
}

synconf() {
  # prepare
  ssh $host mkdir -p $dir
  # remote
  rsync -avP conf.toml $host:$dir
}

deps() {
  go install github.com/swaggo/swag/cmd/swag@latest
  go install github.com/silenceper/gowatch@latest
  go install github.com/gofiber/cli/fiber@latest

  git submodule update --init
  go mod tidy
  run_in_dir "web" pnpm install
}

renewal() {
  local old_name="$pname"
  local new_name="$2"

  if [ -z "$new_name" ]; then
    echo "Missing second parameter."
    exit 0
  fi

  # Define colors for output
  local YELLOW='\033[1;33m'
  local GREEN='\033[0;32m'
  local RED='\033[0;31m'
  local NC='\033[0m'

  echo -e "${YELLOW}[WARNING] Project Initialization & Reset${NC}"
  echo "--------------------------------------------------"
  echo "This script will perform the following IRREVERSIBLE actions:"
  echo -e "1. ${GREEN}GLOBAL REPLACE${NC}: All occurrences of '$old_name' will be changed to '$new_name'."
  echo -e "2. ${RED}GIT RESET${NC}: The existing .git directory will be DELETED."
  echo -e "3. ${YELLOW}RE-INIT${NC}: A new git repository will be initialized in this directory."
  echo -e "4. ${GREEN}PROTECTION${NC}: 'main.sh' will be skipped during string replacement."
  echo "--------------------------------------------------"

  # Prompt for user confirmation
  read -rp "Are you sure you want to proceed? (y/N): " confirm

  # Check user input (defaults to No if empty)
  if [[ "$confirm" =~ ^[yY](es)?$ ]]; then
    # Reset git directory
    rm -rf .git
    rm -rf web/packages/natholdallas
    rm .gitmodules
    # Replace all occurrences of old_name with new_name
    find . \
      \( -name ".git" \
      -o -name "node_modules" \
      -o -name ".nuxt" \
      -o -name ".output" \
      -o -name "dist" \
      -o -name "docs" \
      -o -name "bin" \) -prune \
      -o -type f \
      -exec sed -i "s/${old_name}/${new_name}/g" {} +
    # Initialize git repository
    git init
    git submodule add https://github.com/natholdallas/nuxt-modules.git web/packages/natholdallas
    git add -A
    # Generate docs
    docs
    echo -e "${GREEN}[SUCCESS] Project initialized successfully.${NC}"
  else
    echo -e "${RED}[CANCELLED] Operation aborted by user.${NC}"
    exit 0
  fi
}

case "$1" in
dev) dev "$@" ;;
docs) docs "$@" ;;
build) build "$@" ;;
deploy) deploy "$@" ;;
synconf) synconf ;;
deps) deps "$@" ;;
renewal) renewal "$@" ;;
*)
  echo "Usage:"
  echo "  dev:          Start local development environment (tmux) "
  echo "  docs:         Initialize/Update Swagger documentation "
  echo "  build:        Compile Go backend and generate static sites "
  echo "  deploy:       Build, sync to server, and hot-reload via tmux "
  echo "  synconf:      Sync config to server "
  echo "  deps:         Install dependencies "
  echo "  renewal:      Renewal project "
  exit 1
  ;;
esac
