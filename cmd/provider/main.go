/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	"github.com/crossplane/crossplane-runtime/pkg/external/server"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	themeparkn3wscottcomv1alpha1 "github.com/n3wscott/theme-park-provider/api/v1alpha1"
	"github.com/n3wscott/theme-park-provider/pkg/reconciler/ride"
	"github.com/n3wscott/theme-park-provider/pkg/reconciler/rideoperator"
)

var (
	log logging.Logger
	s   = runtime.NewScheme()
)

func init() {
	// Initialize the scheme with Kubernetes types
	_ = scheme.AddToScheme(s)
	// Add custom API types
	_ = themeparkn3wscottcomv1alpha1.AddToScheme(s)
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize klog flags
	klog.InitFlags(nil)
	if debug {
		_ = flag.Set("v", "5")
	}

	// Initialize logger
	log = logging.NewLogrLogger(textlogger.NewLogger(textlogger.NewConfig()).WithName("theme-park-provider"))

	log.Info("Starting theme park provider")

	// Get gRPC configuration from environment
	grpcEndpoint := os.Getenv("GRPC_ENDPOINT")
	if grpcEndpoint == "" {
		grpcEndpoint = ":50051"
	}
	useTLS := strings.ToLower(os.Getenv("GRPC_USE_TLS")) == "true"
	tlsCertPath := os.Getenv("GRPC_TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("GRPC_TLS_KEY_PATH")

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		log.Info("Received signal", "signal", sig)
		cancel()
	}()

	// Set up the gRPC provider server
	log.Info("Setting up gRPC provider server", "endpoint", grpcEndpoint)

	// Create provider server builder options
	opts := []server.ProviderOption{
		server.WithProviderAddress(grpcEndpoint),
		server.WithProviderLogger(log),
	}

	// Add TLS if enabled
	if useTLS && tlsCertPath != "" && tlsKeyPath != "" {
		opts = append(opts,
			server.WithProviderTLSCertPath(tlsCertPath),
			server.WithProviderTLSKeyPath(tlsKeyPath),
		)
	}

	// Create the provider builder
	builder, err := server.NewProviderBuilder(s, opts...)
	if err != nil {
		log.Info("Failed to create provider builder", "error", err)
		os.Exit(1)
	}

	// Register Ride handler directly with the logger
	if err := builder.RegisterHandler(
		themeparkn3wscottcomv1alpha1.RideGroupVersionKind,
		&ride.ConnectorWrapper{
			Log: log.WithValues("handler", "Ride"),
		},
	); err != nil {
		log.Info("Failed to register Ride handler", "error", err)
		os.Exit(1)
	}

	// Register RideOperator handler directly with the logger
	if err := builder.RegisterHandler(
		themeparkn3wscottcomv1alpha1.RideOperatorGroupVersionKind,
		&rideoperator.ConnectorWrapper{
			Log: log.WithValues("handler", "RideOperator"),
		},
	); err != nil {
		log.Info("Failed to register RideOperator handler", "error", err)
		os.Exit(1)
	}

	// Start the gRPC server
	if err := builder.Start(ctx); err != nil {
		log.Info("Failed to start gRPC server", "error", err)
		os.Exit(1)
	}

	log.Info("gRPC provider server started", "endpoint", grpcEndpoint)

	// Wait for context cancellation
	<-ctx.Done()
	log.Info("Shutting down")
}
