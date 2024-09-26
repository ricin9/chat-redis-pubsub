This is a modern, real-time chat application built with Go, leveraging the power of Fiber, HTMX, Tailwind CSS, and Templ. It uses SQLite for persistent storage and Redis for real-time functionality.

## Features

- Real-time room-based with admin role messaging
- Persistent chat history with cursor pagination
- Persistant unread message count for each room
- Admin can add, kick, promote, demote a member
- Online status for users
- Responsive design with Tailwind CSS
- Server-side rendered templates with Templ
- Client-Server interactions using HTMX

## Tech Stack

- [Go](https://golang.org/) - Backend language
- [Fiber](https://gofiber.io/) - Web framework
- [HTMX](https://htmx.org/) - Frontend interactivity
- [Templ](https://github.com/a-h/templ) - Type-safe Go HTML templating
- [Tailwind CSS](https://tailwindcss.com/) - Styling
- [SQLite](https://www.sqlite.org/) - Database
- [sql-migrate](https://github.com/rubenv/sql-migrate) - Handling database migrations
- [Redis](https://redis.io/) - In-memory data store used for its Pub/Sub capabilities
- [Air](https://github.com/cosmtrek/air) - Live reload for Go apps
- [Docker](https://www.docker.com/) - Containerization
- [Fly.io](https://fly.io/) - Deployment platform

## Prerequisites

- Go 1.20 or higher
- Docker (for deployment)
- Redis server

## Environment Variables

The application uses two environment variables:

- `DATABASE_PATH`: Path to the SQLite database file, Default : `./test.db`
- `REDIS_CONN_STRING`: Connection string for Redis, Default : uses local instance `:6379`

## Development

1. Clone the repository:

```bash
   git clone https://github.com/ricin9/chat-redis-pubsub
   cd chat-redis-pubsub
```

2. Install dependencies:

```bash
   go mod tidy
   go install github.com/rubenv/sql-migrate/...@latest # sql-migrate cli
   go install github.com/air-verse/air@latest          # live reload
   go install github.com/a-h/templ/cmd/templ@latest    # template engine
```

3. Setup redis

You can either install redis locally using your package manager or docker, or use remote connection string like below

```bash
export REDIS_CONN_STRING=redis://<user>:<password>@<host>:<port>
```

4. Run the development server:

I have not yet made development scripts/makefiles yet so you are going to have to do this one by one.

```bash
# runs main.go with live reload
air
```

```bash
# html template generation
templ generate --watch ./views
```

```bash
# generates tailwindcss classes
npx tailwindcss -c ./views/tailwind.config.js -i ./views/input.css -o ./static/css/output.css --watch
```

5. Open your browser and navigate to `http://127.0.0.1:3000`

## Build

1. Generate production templates and styles

```bash
templ generate ./views
npx tailwindcss -c ./views/tailwind.config.js -i ./views/input.css -o ./static/css/output.css --minify
```

2. Build the Docker image

```bash
docker build -t chat-app .
```

## Deployment

This project uses Docker and Fly.io for deployment.

1. Install fly.io CLI if you don't already have it. [Guide](https://fly.io/docs/flyctl/install/)

2. Since fly.io filesystem is ephemeral and we are using sqlite3 db, we need to create a volume.

```bash
fly volume create sqlite_db -r <region> -n 1
```

Region must be the same one you're deploying the app to.

3. Repeat Step 1 from **Build** Steps

4. Deploy to Fly.io:

```bash
flyctl deploy
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Issues

Open an issue if you found a bug, want to sugges changes, refactors, ...etc
