
# Define custom variables for elevator ID and port
ELEVATOR_ID=2
ELEVATOR_PORT=58740

while true; do
    # Open a new terminal window and run the program
    gnome-terminal -- bash -c "./elevator_program -id ${ELEVATOR_ID} -port ${ELEVATOR_PORT}; exec bash"

    # After the terminal window is closed (for example, via Ctrl+C)
    echo "The terminal window was closed."
    echo "Press 'r' to restart the program or 'q' to exit completely:"
    read -n1 choice
    echo
    if [[ $choice == "q" ]]; then
        echo "Exiting..."
        exit 0
    fi
    # If you press 'r', the loop restarts and opens a new terminal window with the same ID and port.
done