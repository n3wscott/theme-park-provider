# Theme Park Provider for Crossplane

This provider allows you to manage theme park resources in Crossplane:
- Rides
- Ride Operators

## Using the Provider

### Prerequisites

- Kubernetes cluster with Crossplane installed
- kubectl
- helm

### Installation

```bash
kubectl apply -f examples/provider.yaml
```

## Simplified Architecture with gRPC

The theme-park-provider now implements a simplified architecture that focuses on gRPC communication between controllers and the provider implementation. The key features are:

1. **Standalone gRPC Server**: The provider runs as a standalone gRPC server without requiring controller-runtime manager.
2. **Direct Resource Management**: Resource handlers connect directly to the gRPC server for operations.
3. **Minimal Dependencies**: Reduced Kubernetes client dependency for running the provider service.
4. **Crossplane Logging**: Uses Crossplane's logging interface for consistent, structured logging.

We still maintain the controller-side setup functions (`SetupWithGRPC`) for backward compatibility with existing controller implementations. These will be removed in future iterations as we move toward a fully decoupled architecture.

### Configuration

Configure the provider using the following environment variables:

```yaml
env:
  - name: GRPC_ENDPOINT
    value: ":50051"  # The address the gRPC server will listen on
  - name: GRPC_USE_TLS
    value: "false"   # Set to "true" to enable TLS
  - name: GRPC_TLS_CERT_PATH
    value: "/certs/tls.crt"  # Path to TLS certificate (when TLS is enabled)
  - name: GRPC_TLS_KEY_PATH
    value: "/certs/tls.key"  # Path to TLS key (when TLS is enabled)
```

### Benefits of the Simplified Architecture

- **Reduced Memory Footprint**: No controller-runtime manager means less memory usage
- **Simplified Deployment**: Running just the gRPC server without full Kubernetes client requirements
- **Clear Separation of Concerns**: Provider logic is cleanly separated from Kubernetes controller logic
- **Network Boundary Crossing**: Controllers and providers can run on different nodes or clusters
- **Scalability**: The server can handle requests from multiple controller instances
- **Security**: Communication can be secured with TLS

### Integration with Crossplane Controllers

When integrating this provider with Crossplane controllers:

1. The controllers connect to the gRPC endpoint specified in the configuration
2. All resource operations (Observe, Create, Update, Delete) flow through the gRPC channel
3. Connection details and resource state are transferred via the gRPC protocol

## Resources

### Ride

The Ride resource represents a theme park ride.

```yaml
apiVersion: themepark.n3wscott.com/v1alpha1
kind: Ride
metadata:
  name: roller-coaster
spec:
  forProvider:
    type: rollercoaster
    capacity: 24
```

### RideOperator

The RideOperator resource represents an operator assigned to a ride.

```yaml
apiVersion: themepark.n3wscott.com/v1alpha1
kind: RideOperator
metadata:
  name: operator-1
spec:
  forProvider:
    ride:
      name: roller-coaster
    frequency: 4
```

## Development

### Building

```bash
make build
```

### Running the Provider

```bash
# Run with default settings
./bin/provider

# Run with custom endpoint
GRPC_ENDPOINT=":8080" ./bin/provider

# Run with TLS enabled
GRPC_USE_TLS=true GRPC_TLS_CERT_PATH="/path/to/cert.crt" GRPC_TLS_KEY_PATH="/path/to/key.key" ./bin/provider
```

### Development with gRPC

When developing new resource types for this provider:

1. Implement the standard ExternalConnector and ExternalClient interfaces
2. Add a ConnectorWrapper that adapts your connector to the gRPC interface
3. Register your handler in main.go

For examples, see:
- `pkg/reconciler/ride/reconciler.go`
- `pkg/reconciler/rideoperator/reconciler.go`

### Connecting Controllers to the Provider

Controllers need to be configured to use the gRPC endpoint:

```go
// Create a connector factory for your resource type
connectorFactory := remote.SetupForResourceType[*v1alpha1.YourResourceType](
    remote.WithEndpoint("your-grpc-endpoint:50051"),
    remote.WithSSL(useTLS),
    remote.WithSetupLogger(logger),
)

// Use the connector in your reconciler
managed.NewReconciler(
    resource.ManagedKind(v1alpha1.YourResourceTypeGroupVersionKind),
    managed.WithTypedExternalConnector(connectorFactory),
)
```

## Implemented Simplifications

We've now removed controller-runtime and the manager from the provider implementation:

### 1. Removed SetupWithGRPC Method

We've eliminated the `SetupWithGRPC` method from each reconciler that was using controller-runtime and creating controllers. This was only needed for the controller side of Crossplane, not for the provider side.

Our standalone gRPC provider now only has:
- The `ConnectorWrapper` implementation
- The `connector` implementation
- The `external` client implementation

### 2. Implemented Direct Registration API

Instead of relying on SetupWith* methods, we now directly register handlers with the gRPC server:

```go
// Direct registration in main.go
builder.RegisterHandler(
    themeparkn3wscottcomv1alpha1.RideGroupVersionKind,
    &ride.ConnectorWrapper{},
)
```

### 3. Removed Controller-Runtime Dependencies

We've reduced controller-runtime dependencies:
- Replaced controller-runtime logging with Crossplane's logging interface
- Removed the controller-runtime client from our ConnectorWrappers
- Implemented direct gRPC server registration without controller-runtime manager

### 4. Next Steps: Package Reorganization

In the future, we could further improve the architecture by reorganizing the code into:
- `pkg/server` - Contains only the provider server implementation
- `pkg/client` - Contains the controller-side client implementation
- `pkg/reconciler` - Contains the shared reconciler logic that both sides use

### Current Implementation

Our current implementation is much simpler and integrates Crossplane's logging interface:

```go
// main.go
func main() {
    // Initialize Crossplane logger
    log = logging.NewLogrLogger(textlogger.NewLogger(textlogger.NewConfig()).WithName("theme-park-provider"))
    
    // Setup gRPC server
    builder, err := remote.NewProviderBuilder(scheme, opts...)
    if err != nil {
        log.Info("Failed to create provider builder", "error", err)
        os.Exit(1)
    }
    
    // Register handlers directly with loggers
    builder.RegisterHandler(
        v1alpha1.RideGroupVersionKind,
        &ride.ConnectorWrapper{
            Log: log.WithValues("handler", "Ride"),
        },
    )
    
    // Handle errors with structured logging
    if err := builder.Start(ctx); err != nil {
        log.Info("Failed to start gRPC server", "error", err)
        os.Exit(1)
    }
    
    log.Info("gRPC provider server started", "endpoint", grpcEndpoint)
}
```

```go
// In reconciler implementation
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
    c.log.Debug("Connecting to provider")
    
    // ...
    
    return &external{log: c.log}, nil
}
```

This implementation is much more lightweight and focused solely on implementing the external resource management logic.

---

## Original Development Notes

Learning how to use the Crossplane Provider API.

Following https://github.com/crossplane/crossplane/blob/main/contributing/guide-provider-development.md

Issues:
 - [ ] Code generation following the guide did not work out as documented.
 - [ ] go.mod tools directive is new and should be used instead of what is documented for `angryjet`
 - [ ] It is harder to not generate the controller with kubebuilder and then wire up the new controller.
 - [ ] A provider is assumed to represent a remote resource but this is not the case for what I happen to be doing.