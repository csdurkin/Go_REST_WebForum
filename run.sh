#!/bin/sh

# PYTHON - Former Implementation
# export FLASK_APP=app.py
# export FLASK_ENV=development
# exec flask run */ 

# GO - Port
# Starts the Go server, already set in main.go with the address and routes
# The Flask version needed FLASK_APP and FLASK_ENV to find and configure the app.
# In other words: main.go provides eveyrthing needed to start server, s run.sh doesn't need to supply anything.
exec go run .