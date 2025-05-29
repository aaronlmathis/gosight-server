package resourcestore

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

type ResourceStore interface {
	// Core CRUD operations
	Create(ctx context.Context, resource *model.Resource) error
	Get(ctx context.Context, id string) (*model.Resource, error)
	Update(ctx context.Context, resource *model.Resource) error
	Delete(ctx context.Context, id string) error

	// Bulk operations
	CreateBatch(ctx context.Context, resources []*model.Resource) error
	UpdateBatch(ctx context.Context, resources []*model.Resource) error

	// Query operations
	List(ctx context.Context, filter *model.ResourceFilter, limit, offset int) ([]*model.Resource, error)
	Count(ctx context.Context, filter *model.ResourceFilter) (int, error)
	Search(ctx context.Context, query *model.ResourceSearchQuery) ([]*model.Resource, error)

	// Relationship queries
	GetChildren(ctx context.Context, parentID string) ([]*model.Resource, error)
	GetParent(ctx context.Context, childID string) (*model.Resource, error)

	// Label/Tag operations
	GetByLabels(ctx context.Context, labels map[string]string) ([]*model.Resource, error)
	GetByTags(ctx context.Context, tags map[string]string) ([]*model.Resource, error)
	UpdateLabels(ctx context.Context, resourceID string, labels map[string]string) error
	UpdateTags(ctx context.Context, resourceID string, tags map[string]string) error

	// Health and status
	UpdateStatus(ctx context.Context, resourceID string, status string) error
	UpdateLastSeen(ctx context.Context, resourceID string, lastSeen time.Time) error
	GetStaleResources(ctx context.Context, threshold time.Duration) ([]*model.Resource, error)

	// Aggregations
	GetResourceSummary(ctx context.Context) (map[string]int, error)
	GetResourcesByKind(ctx context.Context, kind string) ([]*model.Resource, error)
}
