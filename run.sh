# trap 'echo "Ignoring Ctrl+C...";' SIGINT 

# while true; do
#     echo "Building the project..."
#     go build -o elevator_program main.go || { echo "Build failed. Retrying..."; sleep 1; continue; }

#     echo "Starting elevator program..."
#     ./elevator_program

#     echo "Program crashed or terminal closed. Restarting in a new window..."
    
#     sleep 1 
#     gnome-terminal -- bash -c "cd $(pwd); ./run.sh $ID; exec bash"
#     exit 
# done


#!/bin/bash

# Hvis ID ikke er oppgitt, spør brukeren (bruk default 0 hvis enter trykkes)
if [ -z "$1" ]; then
    read -p "Enter elevator ID (default 0): " inputID
    if [ -z "$inputID" ]; then
        ID=0
    else
        ID=$inputID
    fi
else
    ID=$1
fi

# Hvis PORT ikke er oppgitt, spør brukeren (bruk default 15657 hvis enter trykkes)
if [ -z "$2" ]; then
    read -p "Enter elevator port (default 15657): " inputPort
    if [ -z "$inputPort" ]; then
        PORT=15657
    else
        PORT=$inputPort
    fi
else
    PORT=$2
fi

trap 'echo "Ignoring Ctrl+C...";' SIGINT 

while true; do
    echo "Building the project..."
    go build -o elevator_program main.go || { echo "Build failed. Retrying..."; sleep 1; continue; }

    echo "Starting elevator program with ID=$ID and PORT=$PORT..."
    ./elevator_program -id=$ID -port=$PORT

    echo "Program crashed or terminal closed. Restarting in a new window..."
    sleep 1 
    # Åpner nytt terminalvindu og sender med verdiene slik at prompten ikke vises igjen.
    gnome-terminal -- bash -c "cd $(pwd); ./run.sh $ID $PORT; exec bash"
    exit 
done