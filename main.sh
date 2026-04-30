#! /usr/bin/bash

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

  go mod tidy
  run_in_dir "web" pnpm install
}

case "$1" in
dev) dev "$@" ;;
docs) docs ;;
build) build ;;
deploy) deploy "$@" ;;
synconf) synconf ;;
deps) deps ;;
*)
  echo "Usage:"
  echo "  dev:          Start local development environment (tmux) "
  echo "  docs:         Initialize/Update Swagger documentation "
  echo "  build:        Compile Go backend and generate static sites "
  echo "  deploy:       Build, sync to server, and hot-reload via tmux "
  echo "  synconf:      Sync config to server "
  echo "  deps:         Install dependencies "
  exit 1
  ;;
esac
