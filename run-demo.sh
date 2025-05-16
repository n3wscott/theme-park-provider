#\!/bin/bash

# Build both binaries
echo "Building binaries..."
make build-provider
make build-reconciler

# Function to clean up background processes on exit
cleanup() {
  echo "Stopping processes..."
  kill $PROVIDER_PID $RECONCILER_PID 2>/dev/null
  exit
}

# Register cleanup function to run on exit
trap cleanup SIGINT SIGTERM

# Start the provider in its own terminal or background
if [ -x "$(command -v osascript)" ]; then
  # macOS approach
  osascript -e "tell application \"Terminal\" to do script \"cd $(pwd) && export GRPC_ENDPOINT=:50051 && echo 'Starting Provider...' && ./bin/provider --config=${HOME}/.kube/config\""
elif [ -x "$(command -v gnome-terminal)" ]; then
  # Linux with GNOME approach
  gnome-terminal -- bash -c "cd $(pwd) && export GRPC_ENDPOINT=:50051 && echo 'Starting Provider...' && ./bin/provider --config=${HOME}/.kube/config; exec bash"
else
  # Fallback approach - start in background
  echo "Starting Provider in background..."
  GRPC_ENDPOINT=:50051 ./bin/provider --config=${HOME}/.kube/config &
  PROVIDER_PID=$\!
fi

# Give provider time to start
echo "Waiting for provider to start..."
sleep 2

# Start the reconciler in its own terminal or background
if [ -x "$(command -v osascript)" ]; then
  # macOS approach
  osascript -e "tell application \"Terminal\" to do script \"cd $(pwd) && echo 'Starting Reconciler...' && ./bin/reconciler --provider-endpoint=localhost:50051 --config=${HOME}/.kube/config\""
elif [ -x "$(command -v gnome-terminal)" ]; then
  # Linux with GNOME approach
  gnome-terminal -- bash -c "cd $(pwd) && echo 'Starting Reconciler...' && ./bin/reconciler --provider-endpoint=localhost:50051 --config=${HOME}/.kube/config; exec bash"
else
  # Fallback approach - start in background
  echo "Starting Reconciler in background..."
  ./bin/reconciler --provider-endpoint=localhost:50051 --config=${HOME}/.kube/config &
  RECONCILER_PID=$\!
fi

if [[ -n $PROVIDER_PID || -n $RECONCILER_PID ]]; then
  echo "Started in background mode. Press Ctrl+C to stop."
  echo "Provider PID: $PROVIDER_PID"
  echo "Reconciler PID: $RECONCILER_PID"
  # Wait for Ctrl+C
  wait
else
  echo "Started in separate terminals. Close the terminals to stop the processes."
fi
