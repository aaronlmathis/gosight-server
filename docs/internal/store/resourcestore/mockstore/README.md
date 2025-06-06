<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# mockstore

```go
import "github.com/aaronlmathis/gosight-server/internal/store/resourcestore/mockstore"
```

## Index

- [type Call](<#Call>)
- [type CreateCall](<#CreateCall>)
- [type GetCall](<#GetCall>)
- [type MockResourceStore](<#MockResourceStore>)
  - [func NewMockResourceStore\(\) \*MockResourceStore](<#NewMockResourceStore>)
  - [func \(m \*MockResourceStore\) AllCalls\(\) \[\]Call](<#MockResourceStore.AllCalls>)
  - [func \(m \*MockResourceStore\) ClearCalls\(\)](<#MockResourceStore.ClearCalls>)
  - [func \(m \*MockResourceStore\) Count\(ctx context.Context, filter \*model.ResourceFilter\) \(int, error\)](<#MockResourceStore.Count>)
  - [func \(m \*MockResourceStore\) Create\(ctx context.Context, resource \*model.Resource\) error](<#MockResourceStore.Create>)
  - [func \(m \*MockResourceStore\) CreateBatch\(ctx context.Context, resources \[\]\*model.Resource\) error](<#MockResourceStore.CreateBatch>)
  - [func \(m \*MockResourceStore\) CreateCalls\(\) \[\]CreateCall](<#MockResourceStore.CreateCalls>)
  - [func \(m \*MockResourceStore\) Delete\(ctx context.Context, id string\) error](<#MockResourceStore.Delete>)
  - [func \(m \*MockResourceStore\) Get\(ctx context.Context, id string\) \(\*model.Resource, error\)](<#MockResourceStore.Get>)
  - [func \(m \*MockResourceStore\) GetByGroup\(ctx context.Context, group string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetByGroup>)
  - [func \(m \*MockResourceStore\) GetByKind\(ctx context.Context, kind string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetByKind>)
  - [func \(m \*MockResourceStore\) GetByLabels\(ctx context.Context, labels map\[string\]string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetByLabels>)
  - [func \(m \*MockResourceStore\) GetByParent\(ctx context.Context, parentID string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetByParent>)
  - [func \(m \*MockResourceStore\) GetByTags\(ctx context.Context, tags map\[string\]string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetByTags>)
  - [func \(m \*MockResourceStore\) GetCalls\(\) \[\]GetCall](<#MockResourceStore.GetCalls>)
  - [func \(m \*MockResourceStore\) GetChildren\(ctx context.Context, parentID string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetChildren>)
  - [func \(m \*MockResourceStore\) GetParent\(ctx context.Context, childID string\) \(\*model.Resource, error\)](<#MockResourceStore.GetParent>)
  - [func \(m \*MockResourceStore\) GetResourceSummary\(ctx context.Context\) \(map\[string\]int, error\)](<#MockResourceStore.GetResourceSummary>)
  - [func \(m \*MockResourceStore\) GetResourcesByKind\(ctx context.Context, kind string\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetResourcesByKind>)
  - [func \(m \*MockResourceStore\) GetStaleResources\(ctx context.Context, threshold time.Duration\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.GetStaleResources>)
  - [func \(m \*MockResourceStore\) GetStoredResource\(id string\) \(\*model.Resource, bool\)](<#MockResourceStore.GetStoredResource>)
  - [func \(m \*MockResourceStore\) List\(ctx context.Context, filter \*model.ResourceFilter, limit, offset int\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.List>)
  - [func \(m \*MockResourceStore\) Search\(ctx context.Context, query \*model.ResourceSearchQuery\) \(\[\]\*model.Resource, error\)](<#MockResourceStore.Search>)
  - [func \(m \*MockResourceStore\) SetCreateError\(err error\)](<#MockResourceStore.SetCreateError>)
  - [func \(m \*MockResourceStore\) SetGetError\(err error\)](<#MockResourceStore.SetGetError>)
  - [func \(m \*MockResourceStore\) SetUpdateBatchError\(err error\)](<#MockResourceStore.SetUpdateBatchError>)
  - [func \(m \*MockResourceStore\) SetUpdateError\(err error\)](<#MockResourceStore.SetUpdateError>)
  - [func \(m \*MockResourceStore\) Update\(ctx context.Context, resource \*model.Resource\) error](<#MockResourceStore.Update>)
  - [func \(m \*MockResourceStore\) UpdateBatch\(ctx context.Context, resources \[\]\*model.Resource\) error](<#MockResourceStore.UpdateBatch>)
  - [func \(m \*MockResourceStore\) UpdateBatchCalls\(\) \[\]UpdateBatchCall](<#MockResourceStore.UpdateBatchCalls>)
  - [func \(m \*MockResourceStore\) UpdateCalls\(\) \[\]UpdateCall](<#MockResourceStore.UpdateCalls>)
  - [func \(m \*MockResourceStore\) UpdateLabels\(ctx context.Context, resourceID string, labels map\[string\]string\) error](<#MockResourceStore.UpdateLabels>)
  - [func \(m \*MockResourceStore\) UpdateLastSeen\(ctx context.Context, resourceID string, lastSeen time.Time\) error](<#MockResourceStore.UpdateLastSeen>)
  - [func \(m \*MockResourceStore\) UpdateStatus\(ctx context.Context, resourceID string, status string\) error](<#MockResourceStore.UpdateStatus>)
  - [func \(m \*MockResourceStore\) UpdateTags\(ctx context.Context, resourceID string, tags map\[string\]string\) error](<#MockResourceStore.UpdateTags>)
- [type UpdateBatchCall](<#UpdateBatchCall>)
- [type UpdateCall](<#UpdateCall>)


<a name="Call"></a>
## type [Call](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L33-L37>)

Call represents a method call with its arguments

```go
type Call struct {
    Method    string
    Arguments []interface{}
    Timestamp time.Time
}
```

<a name="CreateCall"></a>
## type [CreateCall](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L40-L43>)

CreateCall represents a Create method call

```go
type CreateCall struct {
    Resource *model.Resource
    Error    error
}
```

<a name="GetCall"></a>
## type [GetCall](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L58-L63>)

GetCall represents a Get method call

```go
type GetCall struct {
    ID       string
    Resource *model.Resource
    Found    bool
    Error    error
}
```

<a name="MockResourceStore"></a>
## type [MockResourceStore](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L66-L84>)

MockResourceStore implements resourcestore.ResourceStore for testing

```go
type MockResourceStore struct {
    // contains filtered or unexported fields
}
```

<a name="NewMockResourceStore"></a>
### func [NewMockResourceStore](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L87>)

```go
func NewMockResourceStore() *MockResourceStore
```

NewMockResourceStore creates a new mock resource store

<a name="MockResourceStore.AllCalls"></a>
### func \(\*MockResourceStore\) [AllCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L495>)

```go
func (m *MockResourceStore) AllCalls() []Call
```

AllCalls returns all method calls

<a name="MockResourceStore.ClearCalls"></a>
### func \(\*MockResourceStore\) [ClearCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L505>)

```go
func (m *MockResourceStore) ClearCalls()
```

ClearCalls clears all tracked calls

<a name="MockResourceStore.Count"></a>
### func \(\*MockResourceStore\) [Count](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L545>)

```go
func (m *MockResourceStore) Count(ctx context.Context, filter *model.ResourceFilter) (int, error)
```

Count implements resourcestore.ResourceStore

<a name="MockResourceStore.Create"></a>
### func \(\*MockResourceStore\) [Create](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L99>)

```go
func (m *MockResourceStore) Create(ctx context.Context, resource *model.Resource) error
```

Create implements resourcestore.ResourceStore

<a name="MockResourceStore.CreateBatch"></a>
### func \(\*MockResourceStore\) [CreateBatch](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L236>)

```go
func (m *MockResourceStore) CreateBatch(ctx context.Context, resources []*model.Resource) error
```

CreateBatch implements resourcestore.ResourceStore

<a name="MockResourceStore.CreateCalls"></a>
### func \(\*MockResourceStore\) [CreateCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L465>)

```go
func (m *MockResourceStore) CreateCalls() []CreateCall
```

CreateCalls returns all Create method calls

<a name="MockResourceStore.Delete"></a>
### func \(\*MockResourceStore\) [Delete](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L332>)

```go
func (m *MockResourceStore) Delete(ctx context.Context, id string) error
```

Delete implements resourcestore.ResourceStore

<a name="MockResourceStore.Get"></a>
### func \(\*MockResourceStore\) [Get](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L197>)

```go
func (m *MockResourceStore) Get(ctx context.Context, id string) (*model.Resource, error)
```

Get implements resourcestore.ResourceStore

<a name="MockResourceStore.GetByGroup"></a>
### func \(\*MockResourceStore\) [GetByGroup](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L357>)

```go
func (m *MockResourceStore) GetByGroup(ctx context.Context, group string) ([]*model.Resource, error)
```

GetByGroup implements resourcestore.ResourceStore

<a name="MockResourceStore.GetByKind"></a>
### func \(\*MockResourceStore\) [GetByKind](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L341>)

```go
func (m *MockResourceStore) GetByKind(ctx context.Context, kind string) ([]*model.Resource, error)
```

GetByKind implements resourcestore.ResourceStore

<a name="MockResourceStore.GetByLabels"></a>
### func \(\*MockResourceStore\) [GetByLabels](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L373>)

```go
func (m *MockResourceStore) GetByLabels(ctx context.Context, labels map[string]string) ([]*model.Resource, error)
```

GetByLabels implements resourcestore.ResourceStore

<a name="MockResourceStore.GetByParent"></a>
### func \(\*MockResourceStore\) [GetByParent](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L419>)

```go
func (m *MockResourceStore) GetByParent(ctx context.Context, parentID string) ([]*model.Resource, error)
```

GetByParent implements resourcestore.ResourceStore

<a name="MockResourceStore.GetByTags"></a>
### func \(\*MockResourceStore\) [GetByTags](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L396>)

```go
func (m *MockResourceStore) GetByTags(ctx context.Context, tags map[string]string) ([]*model.Resource, error)
```

GetByTags implements resourcestore.ResourceStore

<a name="MockResourceStore.GetCalls"></a>
### func \(\*MockResourceStore\) [GetCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L455>)

```go
func (m *MockResourceStore) GetCalls() []GetCall
```

GetCalls returns all Get method calls

<a name="MockResourceStore.GetChildren"></a>
### func \(\*MockResourceStore\) [GetChildren](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L624>)

```go
func (m *MockResourceStore) GetChildren(ctx context.Context, parentID string) ([]*model.Resource, error)
```

GetChildren implements resourcestore.ResourceStore

<a name="MockResourceStore.GetParent"></a>
### func \(\*MockResourceStore\) [GetParent](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L629>)

```go
func (m *MockResourceStore) GetParent(ctx context.Context, childID string) (*model.Resource, error)
```

GetParent implements resourcestore.ResourceStore

<a name="MockResourceStore.GetResourceSummary"></a>
### func \(\*MockResourceStore\) [GetResourceSummary](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L722>)

```go
func (m *MockResourceStore) GetResourceSummary(ctx context.Context) (map[string]int, error)
```

GetResourceSummary implements resourcestore.ResourceStore

<a name="MockResourceStore.GetResourcesByKind"></a>
### func \(\*MockResourceStore\) [GetResourcesByKind](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L735>)

```go
func (m *MockResourceStore) GetResourcesByKind(ctx context.Context, kind string) ([]*model.Resource, error)
```

GetResourcesByKind implements resourcestore.ResourceStore

<a name="MockResourceStore.GetStaleResources"></a>
### func \(\*MockResourceStore\) [GetStaleResources](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L435>)

```go
func (m *MockResourceStore) GetStaleResources(ctx context.Context, threshold time.Duration) ([]*model.Resource, error)
```

GetStaleResources implements resourcestore.ResourceStore

<a name="MockResourceStore.GetStoredResource"></a>
### func \(\*MockResourceStore\) [GetStoredResource](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L740>)

```go
func (m *MockResourceStore) GetStoredResource(id string) (*model.Resource, bool)
```

GetStoredResource returns a resource from internal storage \(for testing\)

<a name="MockResourceStore.List"></a>
### func \(\*MockResourceStore\) [List](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L260>)

```go
func (m *MockResourceStore) List(ctx context.Context, filter *model.ResourceFilter, limit, offset int) ([]*model.Resource, error)
```

List implements resourcestore.ResourceStore

<a name="MockResourceStore.Search"></a>
### func \(\*MockResourceStore\) [Search](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L609>)

```go
func (m *MockResourceStore) Search(ctx context.Context, query *model.ResourceSearchQuery) ([]*model.Resource, error)
```

Search implements resourcestore.ResourceStore

<a name="MockResourceStore.SetCreateError"></a>
### func \(\*MockResourceStore\) [SetCreateError](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L517>)

```go
func (m *MockResourceStore) SetCreateError(err error)
```

SetCreateError configures Create to return an error

<a name="MockResourceStore.SetGetError"></a>
### func \(\*MockResourceStore\) [SetGetError](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L538>)

```go
func (m *MockResourceStore) SetGetError(err error)
```

SetGetError configures Get to return an error

<a name="MockResourceStore.SetUpdateBatchError"></a>
### func \(\*MockResourceStore\) [SetUpdateBatchError](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L531>)

```go
func (m *MockResourceStore) SetUpdateBatchError(err error)
```

SetUpdateBatchError configures UpdateBatch to return an error

<a name="MockResourceStore.SetUpdateError"></a>
### func \(\*MockResourceStore\) [SetUpdateError](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L524>)

```go
func (m *MockResourceStore) SetUpdateError(err error)
```

SetUpdateError configures Update to return an error

<a name="MockResourceStore.Update"></a>
### func \(\*MockResourceStore\) [Update](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L131>)

```go
func (m *MockResourceStore) Update(ctx context.Context, resource *model.Resource) error
```

Update implements resourcestore.ResourceStore

<a name="MockResourceStore.UpdateBatch"></a>
### func \(\*MockResourceStore\) [UpdateBatch](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L163>)

```go
func (m *MockResourceStore) UpdateBatch(ctx context.Context, resources []*model.Resource) error
```

UpdateBatch implements resourcestore.ResourceStore

<a name="MockResourceStore.UpdateBatchCalls"></a>
### func \(\*MockResourceStore\) [UpdateBatchCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L485>)

```go
func (m *MockResourceStore) UpdateBatchCalls() []UpdateBatchCall
```

UpdateBatchCalls returns all UpdateBatch method calls

<a name="MockResourceStore.UpdateCalls"></a>
### func \(\*MockResourceStore\) [UpdateCalls](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L475>)

```go
func (m *MockResourceStore) UpdateCalls() []UpdateCall
```

UpdateCalls returns all Update method calls

<a name="MockResourceStore.UpdateLabels"></a>
### func \(\*MockResourceStore\) [UpdateLabels](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L648>)

```go
func (m *MockResourceStore) UpdateLabels(ctx context.Context, resourceID string, labels map[string]string) error
```

UpdateLabels implements resourcestore.ResourceStore

<a name="MockResourceStore.UpdateLastSeen"></a>
### func \(\*MockResourceStore\) [UpdateLastSeen](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L705>)

```go
func (m *MockResourceStore) UpdateLastSeen(ctx context.Context, resourceID string, lastSeen time.Time) error
```

UpdateLastSeen implements resourcestore.ResourceStore

<a name="MockResourceStore.UpdateStatus"></a>
### func \(\*MockResourceStore\) [UpdateStatus](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L688>)

```go
func (m *MockResourceStore) UpdateStatus(ctx context.Context, resourceID string, status string) error
```

UpdateStatus implements resourcestore.ResourceStore

<a name="MockResourceStore.UpdateTags"></a>
### func \(\*MockResourceStore\) [UpdateTags](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L668>)

```go
func (m *MockResourceStore) UpdateTags(ctx context.Context, resourceID string, tags map[string]string) error
```

UpdateTags implements resourcestore.ResourceStore

<a name="UpdateBatchCall"></a>
## type [UpdateBatchCall](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L52-L55>)

UpdateBatchCall represents an UpdateBatch method call

```go
type UpdateBatchCall struct {
    Resources []*model.Resource
    Error     error
}
```

<a name="UpdateCall"></a>
## type [UpdateCall](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/resourcestore/mockstore/mockstore.go#L46-L49>)

UpdateCall represents an Update method call

```go
type UpdateCall struct {
    Resource *model.Resource
    Error    error
}
```

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
