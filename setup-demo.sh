#!/bin/bash

# Build the components
echo "Building components..."
make build

# Install the CRDs
echo "Installing CRDs..."
kubectl apply -f config/crd/bases/

# Create a namespace for our example
echo "Creating namespace..."
kubectl create namespace theme-park-demo 2>/dev/null || true

# Show usage instructions
echo
echo "=================================================="
echo "ThemePark Provider Demo Setup Complete!"
echo "=================================================="
echo 
echo "To run the demo:"
echo "1. Start the provider and reconciler:"
echo "   ./run-demo.sh"
echo
echo "2. In another terminal, create the example resources:"
echo "   kubectl apply -f examples/ride.yaml"
echo "   kubectl apply -f examples/ride-operator.yaml"
echo
echo "3. Check the resources:"
echo "   kubectl get rides"
echo "   kubectl get rideoperators"
echo "   kubectl describe ride roller-coaster"
echo
echo "4. When finished, clean up:"
echo "   kubectl delete -f examples/"
echo "   kubectl delete namespace theme-park-demo"
echo "=================================================="