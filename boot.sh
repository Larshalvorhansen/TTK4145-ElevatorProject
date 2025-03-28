# This script was implemented with significant help from ChatGPT (OpenAI).

if [ -z "$1" ]; then
    echo -e "\nEnter elevator ID (default 0):"
    read -p "> " inputID
    if [ -z "$inputID" ]; then
        ID=0
    else
        ID=$inputID
    fi
else
    ID=$1
fi

if [ -z "$2" ]; then
    echo -e "Enter elevator port (default 15657):"
    read -p "> " inputPort
    if [ -z "$inputPort" ]; then
        PORT=15657
    else
        PORT=$inputPort
    fi
else
    PORT=$2
fi

trap 'echo -e "\nPressed Ctrl+C...";' SIGINT 

while true; do
    echo -e "\nBuilding the project..."
    go build -o restart_program main.go || { echo -e "\nBuild failed. Retrying..."; sleep 1; continue; }

    echo -e "\nStarting elevator program with ID=$ID and PORT=$PORT...\n"
    ./restart_program -id=$ID -port=$PORT

    echo -e "\nProgram crashed or terminal closed. Restarting in a new window...\n"
    sleep 1 
    
    gnome-terminal -- bash -c "cd $(pwd); ./boot.sh $ID $PORT; exec bash"
    exit 
done