#!/bin/bash

# Use first parameter as ID and second as PORT (default to 58740 if not provided)
ID=$1
PORT=${2:-58740}

if [ -z "$ID" ]; then
    echo "Usage: ./run.sh <ID> [PORT]"
    exit 1
fi

# Ignore Ctrl+C so the script loop is not interrupted
trap 'echo "Ignoring Ctrl+C...";' SIGINT

while true; do
    echo "Building the project..."
    go build -o elevator main.go || { echo "Build failed. Retrying..."; sleep 1; continue; }

    echo "Starting elevator program with ID=$ID and PORT=$PORT..."
    ./elevator -id=$ID -port=$PORT

    echo "Program crashed or terminal closed. Restarting in a new window..."
    sleep 1 
    # Opens a new terminal, changes to the current directory, and runs the script with the same parameters
    gnome-terminal -- bash -c "cd $(pwd); ./run.sh $ID $PORT; exec bash"
    exit 
done