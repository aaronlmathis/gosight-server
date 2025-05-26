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

// server/internal/http/httpserver.go
// Basic http server for admin/dash

package httpserver

import (
	"net/http"

	gosighttemplate "github.com/aaronlmathis/gosight-server/internal/http/templates"
	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/gorilla/mux"
)

type HttpServer struct {
	httpServer *http.Server
	Router     *mux.Router
	Sys        *sys.SystemContext
	Tmpl       *gosighttemplate.GoSightTemplate
}

// NewServer creates a new HTTP server instance with the provided system context.
// It initializes the router and sets up the server configuration.
func NewServer(sys *sys.SystemContext) *HttpServer {
	router := mux.NewRouter()
	router.StrictSlash(true)
	s := &HttpServer{
		httpServer: &http.Server{
			Addr:    sys.Cfg.Server.HTTPAddr,
			Handler: router,
		},
		Router: router,
		Sys:    sys,
	}

	return s
}

// Start initializes the HTTP server, sets up routes, and starts listening for incoming requests.
func (s *HttpServer) Start() error {

	s.setupRoutes()

	tmpl, err := gosighttemplate.NewGoSightTemplate(s.Sys.Ctx, s.Sys.Cfg, s.Sys.Stores.Metrics, s.Sys.Tele.Index, s.Sys.Stores.Users)
	if err != nil {
		utils.Fatal("template init failed: %v", err)
	}

	s.Tmpl = tmpl

	utils.Info("HTTPS server running at %s", s.Sys.Cfg.Server.HTTPAddr)
	if err := http.ListenAndServeTLS(s.Sys.Cfg.Server.HTTPAddr, s.Sys.Cfg.TLS.HttpsCertFile, s.Sys.Cfg.TLS.HttpsKeyFile, s.Router); err != nil {
		utils.Error("HTTPS server failed: %v", err)
		return err
	}
	return nil
}

// Shutdown gracefully stops the HTTP server, allowing for cleanup and ensuring no active connections are abruptly terminated.
func (s *HttpServer) Shutdown() error {
	utils.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(s.Sys.Ctx); err != nil {
		utils.Error("HTTP shutdown error: %v", err)
		return err
	}

	utils.Info("HTTP server shut down cleanly")
	return nil
}
