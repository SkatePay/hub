#!/bin/sh

# Start the first service (API)
./hub-service api &

# Start the second service (Chatbot)
./hub-service chatbot &

# Start the third service (Broadcast)
./hub-service support &

# Wait for all background processes to complete
wait