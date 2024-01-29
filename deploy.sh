#!/bin/sh

source .env

# Push container to registry
heroku container:push web

# Release from register
heroku container:release web

# Optional: Test Heroku endpoint using websocat 
websocat wss://$HEROKU_URL