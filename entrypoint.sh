#!/bin/sh

# Start the first service (API)
./hub-service api

# Wait for all background processes to complete
wait