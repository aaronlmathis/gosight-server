/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// Package main is the package for the GoSight server.
// It initializes the server, sets up the gRPC and HTTP servers, and handles graceful shutdown.
// It also manages the system context and sync manager for periodic persistence.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aaronlmathis/gosight-server/internal/bootstrap"
	grpcserver "github.com/aaronlmathis/gosight-server/internal/grpc"
	httpserver "github.com/aaronlmathis/gosight-server/internal/http"
	"github.com/aaronlmathis/gosight-server/internal/otel"
	"github.com/aaronlmathis/gosight-server/internal/syslog"
	"github.com/aaronlmathis/gosight-shared/utils"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"
)

// run is the initialization function for the GoSight server.
// It initializes the server, sets up the gRPC and HTTP servers, and handles graceful shutdown.
func run(configFlag *string) {
	// Graceful Shutdown Context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh
		utils.Warn("Received signal: %s", sig)

		cancel()

	}()

	// Init System Context for Gosight
	sys, err := bootstrap.InitGoSight(ctx, configFlag)
	utils.Must("System Context", err)

	// Start SyncManager for periodic persistence
	// Start all sync loops — this blocks until ctx is canceled
	go sys.SyncMgr.Run()

	// Start HTTP server for admin console/api
	srv := httpserver.NewServer(sys)

	go func() {
		if err = srv.Start(); err != nil {
			utils.Fatal("HTTP server failed: %v", err)
		} else {
			utils.Info("HTTP server started successfully")
		}
	}()

	// Start OTel receiver for metrics and logs
	// TODO: GRPC?
	otelReceiver, err := otel.NewOTelReceiver(sys)
	if sys.Cfg.OpenTelemetry.HTTP.Enabled {
		if err != nil {
			utils.Fatal("Failed to create OTel receiver: %v", err)
		}
		go func() {
			if err := otelReceiver.Start(); err != nil {
				utils.Fatal("Failed to start OTel receiver: %v", err)
			} else {
				utils.Info("OTel receiver started successfully: listening on %s", sys.Cfg.OpenTelemetry.HTTP.Addr)
			}
		}()
	}

	// Start Syslog server
	syslogServer, err := syslog.NewSyslogServer(sys)
	utils.Must("Syslog", err)
	go func() {
		if err := syslogServer.Start(); err != nil {
			utils.Fatal("Syslog server failed: %v", err)
		} else {
			utils.Info("Syslog server started successfully: listening on TCP: %d, UDP: %d", sys.Cfg.SyslogCollection.TCPPort, sys.Cfg.SyslogCollection.UDPPort)
		}
	}()

	// register gzip codec for compression
	_ = gzip.Name // This ensures the gzip codec is registered
	utils.Debug("Log store is: %T", sys.Stores.Logs)
	grpcServer, err := grpcserver.NewGRPCServer(sys)

	if err != nil {
		utils.Fatal("Failed to start gRPC server: %v", err)
	} else {
		go func() {
			utils.Info("GoSight server listening on %s", sys.Cfg.Server.GRPCAddr)
			if err := grpcServer.Server.Serve(grpcServer.Listener); err != nil {
				utils.Fatal("Failed to serve gRPC: %v", err)
			}
		}()
	}

	<-ctx.Done()
	utils.Info("Shutting down GoSight...")

	if err := srv.Shutdown(); err != nil {
		utils.Warn("Failed to shutdown HTTP server: %v", err)
	}

	if err := otelReceiver.Shutdown(); err != nil {
		utils.Warn("Failed to shutdown OTel receiver: %v", err)
	}
	// Tell Agents to disconnect gracefully
	grpcServer.GracefulDisconnectAllAgents()
	// Close GRPC gracefully.
	grpcServer.Server.GracefulStop()

	// Stop Syslog server gracefully.
	syslogServer.Stop()

	// Flush all pending data before shutdown.

	// Stop resource cache to ensure final flush of dirty resources
	sys.Cache.Resources.Stop()

	// Disconnect from metric store, datastore, and userstore.
	if err := sys.Stores.Metrics.Close(); err != nil {
		utils.Warn("Failed to close metric store: %v", err)
	}
	if err := sys.Stores.Data.Close(); err != nil {
		utils.Warn("Failed to close datastore: %v", err)
	}
	if err := sys.Stores.Users.Close(); err != nil {
		utils.Warn("Failed to close userstore: %v", err)
	}
}

// main is the entry point for the GoSight server.
func main() {
	versionFlag := flag.Bool("version", false, "print version information and exit")
	configFlag := flag.String("config", "", "Path to server config file")
	flag.Parse()
	if *versionFlag {
		fmt.Printf(
			"GoSight %s (built %s, commit %s)\n",
			Version, BuildTime, GitCommit,
		)
		os.Exit(0)
	}
	run(configFlag)
}
