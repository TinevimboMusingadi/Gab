package main

import (
	"context"
	"time"

	"gab/internal/models"
	"gab/internal/query"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Note: This file will need the generated gRPC code from the proto file
// For now, we'll create placeholder implementations
// To generate: protoc --go_out=. --go-grpc_out=. api/proto/gab.proto

// RegisterGabService registers the Gab service with gRPC server
// This is a placeholder - actual implementation requires generated code
func RegisterGabService(s *grpc.Server, server *Server) {
	// Implementation will be:
	// gabv1.RegisterGabServer(s, &gabService{server: server})
}

// gabService implements the Gab gRPC service
type gabService struct {
	// UnimplementedGabServer must be embedded for forward compatibility
	// UnimplementedGabServer
	server *Server
}

// ListAgents returns all agent IDs
func (s *gabService) ListAgents(ctx context.Context, req *ListAgentsRequest) (*ListAgentsResponse, error) {
	agents, err := s.server.queryEngine.GetAllAgents()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &ListAgentsResponse{AgentIds: agents}, nil
}

// ListEvents returns paginated events for an agent
func (s *gabService) ListEvents(ctx context.Context, req *ListEventsRequest) (*ListEventsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)

	chain, err := s.server.queryEngine.GetEventChain(req.AgentId, limit+offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	events := chain.Events
	if offset < len(events) {
		if offset+limit > len(events) {
			events = events[offset:]
		} else {
			events = events[offset : offset+limit]
		}
	} else {
		events = []*models.Event{}
	}

	// Convert to proto events (placeholder - need generated types)
	// protoEvents := make([]*gabv1.Event, len(events))
	// for i, ev := range events {
	//     protoEvents[i] = convertEventToProto(ev)
	// }

	return &ListEventsResponse{
		// Events: protoEvents,
		Total: int32(chain.Count),
	}, nil
}

// GetEventChain returns the event chain for an agent
func (s *gabService) GetEventChain(ctx context.Context, req *GetEventChainRequest) (*GetEventChainResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}

	chain, err := s.server.queryEngine.GetEventChain(req.AgentId, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto events
	// protoEvents := make([]*gabv1.Event, len(chain.Events))
	// for i, ev := range chain.Events {
	//     protoEvents[i] = convertEventToProto(ev)
	// }

	return &GetEventChainResponse{
		// Events: protoEvents,
	}, nil
}

// GetStateDiff returns the difference between two states
func (s *gabService) GetStateDiff(ctx context.Context, req *GetStateDiffRequest) (*GetStateDiffResponse, error) {
	diff, err := s.server.diffEngine.GetStateDiff(req.PreStateId, req.PostStateId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto format
	created := make([]*FileChange, len(diff.Created))
	for i, c := range diff.Created {
		created[i] = &FileChange{
			Path:          c.Path,
			PostArtifactId: c.PostArtifactID,
		}
	}

	modified := make([]*FileChange, len(diff.Modified))
	for i, m := range diff.Modified {
		modified[i] = &FileChange{
			Path:           m.Path,
			PreArtifactId:  m.PreArtifactID,
			PostArtifactId: m.PostArtifactID,
		}
	}

	deleted := make([]*FileChange, len(diff.Deleted))
	for i, d := range diff.Deleted {
		deleted[i] = &FileChange{
			Path:          d.Path,
			PreArtifactId: d.PreArtifactID,
		}
	}

	return &GetStateDiffResponse{
		Created:  created,
		Modified: modified,
		Deleted:  deleted,
	}, nil
}

// QueryEvents performs an advanced query
func (s *gabService) QueryEvents(ctx context.Context, req *QueryEventsRequest) (*QueryEventsResponse, error) {
	q := s.server.queryEngine.NewQuery()
	if req.AgentId != "" {
		q.Agent(req.AgentId)
	}
	if req.ActionType != "" {
		q.ActionType(req.ActionType)
	}
	if req.Tool != "" {
		q.Tool(req.Tool)
	}
	if req.StartTime > 0 && req.EndTime > 0 {
		q.TimeRange(time.Unix(req.StartTime, 0), time.Unix(req.EndTime, 0))
	}
	if req.Limit > 0 {
		q.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		q.Offset(int(req.Offset))
	}

	events, err := q.Execute()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	total, _ := q.Count()

	// Convert to proto events
	// protoEvents := make([]*gabv1.Event, len(events))
	// for i, ev := range events {
	//     protoEvents[i] = convertEventToProto(ev)
	// }

	return &QueryEventsResponse{
		// Events: protoEvents,
		Total: int32(total),
	}, nil
}

// GetArtifact returns an artifact by ID
func (s *gabService) GetArtifact(ctx context.Context, req *GetArtifactRequest) (*GetArtifactResponse, error) {
	var art models.Artifact
	if err := s.server.store.GetObject(req.ArtifactId, &art); err != nil {
		return nil, status.Error(codes.NotFound, "artifact not found")
	}

	// Convert to proto
	// protoArt := &gabv1.Artifact{
	//     Id:          art.ID,
	//     ContentType: art.ContentType,
	//     Data:        art.Data,
	// }

	return &GetArtifactResponse{
		// Artifact: protoArt,
	}, nil
}

// Placeholder types (will be replaced by generated proto code)
type ListAgentsRequest struct{}
type ListAgentsResponse struct {
	AgentIds []string
}
type ListEventsRequest struct {
	AgentId string
	Limit   int32
	Offset  int32
}
type ListEventsResponse struct {
	Events []interface{} // *gabv1.Event
	Total  int32
}
type GetEventChainRequest struct {
	AgentId string
	Limit   int32
}
type GetEventChainResponse struct {
	Events []interface{} // *gabv1.Event
}
type GetStateDiffRequest struct {
	PreStateId  string
	PostStateId string
}
type GetStateDiffResponse struct {
	Created  []*FileChange
	Modified []*FileChange
	Deleted  []*FileChange
}
type FileChange struct {
	Path          string
	PreArtifactId  string
	PostArtifactId string
}
type QueryEventsRequest struct {
	AgentId    string
	ActionType string
	Tool       string
	StartTime  int64
	EndTime    int64
	Limit      int32
	Offset     int32
}
type QueryEventsResponse struct {
	Events []interface{} // *gabv1.Event
	Total  int32
}
type GetArtifactRequest struct {
	ArtifactId string
}
type GetArtifactResponse struct {
	Artifact interface{} // *gabv1.Artifact
}

