# Web Template

A full-stack web application template built with Go (Fiber), Vue.js, and MySQL.

## Tech Stack

| Layer    | Technology           |
| -------- | -------------------- |
| Backend  | Go 1.25 + Fiber v3   |
| ORM      | GORM + MySQL/Mariadb |
| Cache    | Redis                |
| Frontend | Vue 3 + Nuxt         |
| Config   | Viper                |

## Features

- Multi-app architecture (admin & user portals)
- JWT authentication
- RESTful API with Swagger docs
- Hot reload development
- Production build & deployment scripts

## Quick Start

```bash
# Development
./main.sh dev

# Build
./main.sh build

# Generate API docs
./main.sh docs

# Deploy
./main.sh deploy
```

## Structure

```
├── main.go             # Entry point
├── conf.toml           # Configuration
├── internal/           # Backend code
│   ├── srv/            # HTTP handlers
│   ├── db/             # Database models
│   └── task/           # Background tasks
└── web/                # Frontend apps
    ├── apps/admin/     # Admin portal
    └── apps/user/      # User portal
```

## License

MIT
