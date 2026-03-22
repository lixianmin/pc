#!/bin/bash

echo "Stopping old pc process..."
./pc gateway stop
sleep 1

echo "Building pc..."
make build

echo "Building plugins..."
make build-plugins

echo "Starting pc..."
./pc gateway start


sleep 1
./pc tui
