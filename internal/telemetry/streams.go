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

// gosight/server/internal/api
// streams.go - gRPC handler for stream submission

package telemetry

import (
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	pb "github.com/aaronlmathis/gosight-shared/proto"
	"github.com/aaronlmathis/gosight-shared/utils"
	"google.golang.org/protobuf/proto"
)

// StreamHandler implements pb.MetricsServiceServer
// StreamHandler implements MetricsServiceServer
type StreamHandler struct {
	Sys *sys.SystemContext
	pb.UnimplementedStreamServiceServer
}

func NewStreamHandler(sys *sys.SystemContext) *StreamHandler {
	utils.Debug("StreamHandler initialized with store: %T", sys.Stores.Metrics)
	return &StreamHandler{
		Sys: sys,
	}
}

// Stream implements the gRPC StreamService_StreamServer method
func (h *StreamHandler) Stream(stream pb.StreamService_StreamServer) error {
	var agentID string

	go func() {
		for {
			time.Sleep(5 * time.Second)
			//utils.Debug("server Stream handler alive...")
		}
	}()

	for {

		req, err := stream.Recv()
		if err != nil {
			utils.Error("Stream receive error: %v", err)
			return err
		}

		switch v := req.Payload.(type) {

		case *pb.StreamPayload_Process:
			SafeHandlePayload(func() {
				var processPayload pb.ProcessPayload
				if err := proto.Unmarshal(v.Process.RawPayload, &processPayload); err != nil {
					utils.Warn("Failed to unmarshal ProcessPayload: %v", err)
				}
				// Convert the protobuf ProcessPayload to model.ProcessPayload
				converted := ConvertProtoProcessPayload(&processPayload)

				// Tag enrichment
				if converted.Meta != nil && converted.Meta.EndpointID != "" {
					tags, err := h.Sys.Stores.Data.GetTags(stream.Context(), converted.Meta.EndpointID)
					if err == nil {
						if converted.Meta.Tags == nil {
							converted.Meta.Tags = make(map[string]string)
						}
						for k, v := range tags {
							if _, exists := converted.Meta.Tags[k]; !exists {
								converted.Meta.Tags[k] = v
							}
						}
					}
				}

				h.Sys.Cache.Processes.Add(model.ProcessSnapshot{
					Timestamp:  converted.Timestamp,
					HostID:     converted.HostID,
					EndpointID: converted.EndpointID,
					Processes:  converted.Processes,
					Meta:       converted.Meta,
				})

				// Broadcast + store process payload

				h.Sys.WSHub.Processes.Broadcast(converted)

				// Write Process snapshots to the buffer datastore

				if err := h.Sys.Buffers.Data.WriteAny(converted); err != nil {
					// Insert Process Snapshot and ProcessInfos into database.
					if err := h.Sys.Stores.Data.InsertFullProcessPayload(stream.Context(), &converted); err != nil {
						utils.Warn("Failed to store ProcessPayload: %v", err)
						return
					} else {
						//utils.Debug("Stored ProcessPayload from agent %s", converted.EndpointID)
					}
				}

			})

		case *pb.StreamPayload_CommandResponse:
			// Received command execution result
			//utils.Info("Received CommandResponse: success=%v output=%s error=%s",	v.CommandResponse.Success, v.CommandResponse.Output, v.CommandResponse.ErrorMessage)

			if endpointID, ok := h.Sys.Tracker.GetEndpointIdByAgentId(agentID); ok {
				h.Sys.WSHub.Commands.Broadcast(&model.CommandResult{

					EndpointID:   endpointID,
					Output:       v.CommandResponse.Output,
					Success:      v.CommandResponse.Success,
					ErrorMessage: v.CommandResponse.ErrorMessage,
					Timestamp:    time.Now().Format(time.RFC3339),
				})
			} else {
				utils.Warn("No endpoint ID found for agent %s", agentID)
			}

		case *pb.StreamPayload_CommandRequest:
			// Server shouldn't receive CommandRequests — log it
			utils.Warn("Unexpected CommandRequest received from agent")

		default:
			utils.Warn("Unknown payload type received on stream")
		}

		//  Always send a StreamResponse after processing
		resp := &pb.StreamResponse{
			Status:     "ok",
			StatusCode: 0,
		}

		//utils.Debug("Agent ID: %s", agentID)
		var pendingCmd *pb.CommandRequest
		if agentID != "" {
			pendingCmd = h.Sys.Tracker.DequeueCommand(agentID) // Copy command reference outside of lock
		}
		if pendingCmd != nil {
			resp.Command = pendingCmd
			utils.Info("Injecting pending CommandRequest into StreamResponse for agent %s", agentID)
		}
		//utils.Debug("Sending StreamResponse to %s", agentID)
		if session, ok := h.Sys.Tracker.GetAgentSession(agentID); ok {
			select {
			case session.SendQueue <- resp:
				// sent successfully
			default:
				utils.Warn("SendQueue full for agent %s — dropping response", agentID)
			}
		} else {
			//utils.Warn("No live session found for agent %s", agentID)
		}
	}
}
