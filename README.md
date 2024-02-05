# GGJ_Platform
Platform code for Global Game Jam 2024.

```
# Start the server
go run cmd/server/main.go -env dev -verbose
```

## Environment Setup
`.env` files are used to define the environment for the websocket server and are required to deploy/test the server in production.

```
# config/.env

# Port to run the local server on
PORT=

# URL for the Heroku dyno
HEROKU_URL=
```