<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# routestore

```go
import "github.com/aaronlmathis/gosight-server/internal/store/routestore"
```

SPDX\-License\-Identifier: GPL\-3.0\-or\-later

Copyright \(C\) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or \(at your option\) any later version.

GoSight is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with GoSight. If not, see https://www.gnu.org/licenses/.

Package routestore provides functionality to load and manage action routes

## Index

- [type RouteStore](<#RouteStore>)
  - [func NewRouteStore\(path string\) \(\*RouteStore, error\)](<#NewRouteStore>)
  - [func \(rs \*RouteStore\) BuildMap\(\) map\[string\]model.ActionRoute](<#RouteStore.BuildMap>)


<a name="RouteStore"></a>
## type [RouteStore](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/routestore/loader.go#L32-L34>)

RouteStore is a structure that holds a list of action routes.

```go
type RouteStore struct {
    Routes []model.ActionRoute
}
```

<a name="NewRouteStore"></a>
### func [NewRouteStore](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/routestore/loader.go#L37>)

```go
func NewRouteStore(path string) (*RouteStore, error)
```

LoadRoutesFromFile loads action routes from a YAML file.

<a name="RouteStore.BuildMap"></a>
### func \(\*RouteStore\) [BuildMap](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/routestore/loader.go#L51>)

```go
func (rs *RouteStore) BuildMap() map[string]model.ActionRoute
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
