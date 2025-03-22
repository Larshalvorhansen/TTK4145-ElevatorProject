trap 'echo "Ignoring Ctrl+C...";' SIGINT 

while true; do
    echo "Building the project..."
    go build -o elevator_program main.go || { echo "Build failed. Retrying..."; sleep 1; continue; }

    echo "Starting elevator program..."
    ./elevator_program

    echo "Program crashed or terminal closed. Restarting in a new window..."
    
    sleep 1 
    gnome-terminal -- bash -c "cd $(pwd); ./run.sh $ID; exec bash"
    exit 
done
