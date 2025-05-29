// Package userstore defines the RBAC storage interface and types
package userstore

import (
	"context"

	"github.com/aaronlmathis/gosight-server/internal/usermodel"
)

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error)
	GetUserByUsername(ctx context.Context, username string) (*usermodel.User, error)
	GetUserByID(ctx context.Context, username string) (*usermodel.User, error)
	GetUserWithPermissions(ctx context.Context, userID string) (*usermodel.User, error)
	GetAllUsersWithPermissions(ctx context.Context) ([]*usermodel.User, error)
	GetUserBySSO(ctx context.Context, provider string, ssoID string) (*usermodel.User, error)
	SaveUser(ctx context.Context, u *usermodel.User) error
	CreateUser(ctx context.Context, u *usermodel.User) error
	UpdateUser(ctx context.Context, u *usermodel.User) error
	DeleteUser(ctx context.Context, userID string) error
	// Role management
	GetRoles(ctx context.Context) ([]*usermodel.Role, error)
	GetRole(ctx context.Context, roleID string) (*usermodel.Role, error)
	CreateRole(ctx context.Context, r *usermodel.Role) error
	UpdateRole(ctx context.Context, r *usermodel.Role) error
	DeleteRole(ctx context.Context, roleID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*usermodel.Permission, error)
	AssignPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error
	GetUsersWithRole(ctx context.Context, roleID string) ([]*usermodel.User, error)

	// Permission management
	GetPermissions(ctx context.Context) ([]*usermodel.Permission, error)
	GetPermission(ctx context.Context, permissionID string) (*usermodel.Permission, error)
	CreatePermission(ctx context.Context, p *usermodel.Permission) error
	UpdatePermission(ctx context.Context, p *usermodel.Permission) error
	DeletePermission(ctx context.Context, permissionID string) error
	GetRolesWithPermission(ctx context.Context, permissionID string) ([]*usermodel.Role, error)

	// User role management
	GetUserRoles(ctx context.Context, userID string) ([]*usermodel.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	AssignRolesToUser(ctx context.Context, userID string, roleIDs []string) error
	RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error
	AttachPermissionToRole(ctx context.Context, roleID, permID string) error
	Close() error

	// Profile management
	GetUserProfile(ctx context.Context, userID string) (*usermodel.UserProfile, error)
	CreateUserProfile(ctx context.Context, profile *usermodel.UserProfile) error
	UpdateUserProfile(ctx context.Context, profile *usermodel.UserProfile) error

	// Settings management
	GetUserSettings(ctx context.Context, userID string) (usermodel.UserSettings, error)
	GetUserSetting(ctx context.Context, userID, key string) (*usermodel.UserSetting, error)
	SetUserSetting(ctx context.Context, userID, key string, value []byte) error
	DeleteUserSetting(ctx context.Context, userID, key string) error

	// Password management
	UpdateUserPassword(ctx context.Context, userID, passwordHash string) error

	// Complete user operations
	GetCompleteUser(ctx context.Context, userID string) (*usermodel.CompleteUser, error)
}
