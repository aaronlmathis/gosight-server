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

package syslog

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

/*
// NetworkDevice represents a syslog-emitting network appliance

	type NetworkDevice struct {
		ID        string    `db:"id"`         // UUID primary key
		Name      string    `db:"name"`       // human-readable label
		Vendor    string    `db:"vendor"`     // e.g. "sonicwall", "fortinet"
		Address   string    `db:"address"`    // IP or hostname
		Port      int       `db:"port"`       // syslog port (default 514)
		Protocol  string    `db:"protocol"`   // "udp", "tcp"
		Format    string    `db:"format"`     // "rfc3164", "rfc5424", "cef", etc.
		Facility  string    `db:"facility"`   // syslog facility, e.g. "local0"
		SyslogID  string    `db:"syslog_id"`  // vendor tag or hostname override
		RateLimit int       `db:"rate_limit"` // events/sec throttle (optional)
		CreatedAt time.Time `db:"created_at"` // record creation time
		UpdatedAt time.Time `db:"updated_at"` // record update time
	}
*/
type SyslogServer struct {
	listener          net.Listener
	sys               *sys.SystemContext
	wg                sync.WaitGroup
	NetworkDevices    []*model.NetworkDevice
	maxConnections    int
	activeConnections int
	connectionMutex   sync.Mutex
	ipConnections     map[string]int
	ipLimits          map[string]int // Optional per-IP limits
	defaultIPLimit    int
}

// NewSyslogServer creates a new SyslogServer instance
func NewSyslogServer(sys *sys.SystemContext) (*SyslogServer, error) {

	networkDevices, err := sys.Stores.Data.GetNetworkDevices(sys.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network devices: %w", err)
	}

	var maxConnections, defaultIPLimit int
	if sys.Cfg.SyslogCollection.MaxConnections < 1 {
		maxConnections = 500
	} else {
		maxConnections = sys.Cfg.SyslogCollection.MaxConnections
	}

	if sys.Cfg.SyslogCollection.DefaultIPLimit < 1 {
		defaultIPLimit = 10
	} else {
		defaultIPLimit = sys.Cfg.SyslogCollection.DefaultIPLimit
	}

	return &SyslogServer{
		sys:               sys,
		NetworkDevices:    networkDevices,
		maxConnections:    maxConnections,
		activeConnections: 0,
		ipConnections:     make(map[string]int),
		ipLimits:          make(map[string]int),
		defaultIPLimit:    defaultIPLimit,
	}, nil
}

// Start binds sockets, wires context cancellation, then spins up the two listeners.
func (s *SyslogServer) Start() error {

	// Bind sockets
	udpConn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", s.sys.Cfg.SyslogCollection.UDPPort))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	tcpLn, err := net.Listen("tcp", fmt.Sprintf(":%d", s.sys.Cfg.SyslogCollection.TCPPort))
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to listen on TCP: %w", err)
	}

	// When context is done, close both sockets to unblock Read/Accept
	go func() {
		<-s.sys.Ctx.Done()
		udpConn.Close()
		tcpLn.Close()
	}()

	// Launch UDP listener
	s.wg.Add(1)
	go s.listenUDP(s.sys.Ctx, udpConn)

	// Launch TCP listener
	s.wg.Add(1)
	go s.listenTCP(s.sys.Ctx, tcpLn)

	return nil
}

// isAuthorizedIP checks if the given IP address matches any of our configured network devices
func (s *SyslogServer) isAuthorizedIP(ip string) bool {
	utils.Debug("Checking if %s is authorized", ip)
	for _, device := range s.NetworkDevices {
		if device.Address == ip {
			return true
		}
	}
	return false
}

// listenUDP reads packets until ctx is canceled or the socket is closed.
func (s *SyslogServer) listenUDP(ctx context.Context, conn net.PacketConn) {
	defer s.wg.Done()

	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		// Extract IP without port
		srcStr := src.String()
		ip := srcStr
		if host, _, err := net.SplitHostPort(srcStr); err == nil {
			ip = host
		}

		// Check if the IP is authorized
		if !s.isAuthorizedIP(ip) {
			utils.Warn("Unauthorized syslog UDP connection attempt from %s", ip)
			continue
		}

		// Handle the raw syslog packet
		go s.handleLog(ctx, data, ip)
	}
}

// listenTCP accepts connections and reads lines until ctx is canceled or the listener is closed.
func (s *SyslogServer) listenTCP(ctx context.Context, ln net.Listener) {
	defer s.wg.Done()

	for {
		// Check if we can accept more connections
		s.connectionMutex.Lock()
		if s.activeConnections >= s.maxConnections {
			s.connectionMutex.Unlock()
			// Wait before trying again
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
		s.activeConnections++
		s.connectionMutex.Unlock()

		// Accept with timeout to prevent blocking forever
		deadline := time.Now().Add(5 * time.Second)
		ln.(*net.TCPListener).SetDeadline(deadline)

		conn, err := ln.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Just a timeout, try again
			}
			// Other errors mean listener closed or fatal error
			s.connectionMutex.Lock()
			s.activeConnections--
			s.connectionMutex.Unlock()
			return
		}

		ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			utils.Error("Failed to parse remote address: %v", err)
			conn.Close()
			continue
		}

		// Check if the IP is authorized
		if !s.isAuthorizedIP(ip) {
			utils.Warn("Unauthorized syslog TCP connection attempt from %s", ip)
			conn.Close()
			continue
		}

		s.connectionMutex.Lock()
		ipCount := s.ipConnections[ip]
		ipLimit := s.defaultIPLimit
		if limit, ok := s.ipLimits[ip]; ok {
			ipLimit = limit
		}

		if ipCount >= ipLimit {
			s.connectionMutex.Unlock()
			conn.Close() // Reject connection
			continue
		}

		s.ipConnections[ip]++
		s.connectionMutex.Unlock()

		go s.handleTCPConn(ctx, conn)
	}
}

// handleTCPConn reads syslog lines from a single connection.
func (s *SyslogServer) handleTCPConn(ctx context.Context, conn net.Conn) {
	defer func() {
		conn.Close()
		s.connectionMutex.Lock()
		s.activeConnections--
		s.connectionMutex.Unlock()
	}()

	// Set timeout for idle connections
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		// Reset deadline after successful read
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Wire in the handleLog helper
		go s.handleLog(ctx, line, conn.RemoteAddr().String())
	}
}

// Stop triggers a shutdown (via context cancellation upstream) and waits for both listeners.
func (s *SyslogServer) Stop() {

	s.wg.Wait()
}
