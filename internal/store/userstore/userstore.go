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
	CreateRole(ctx context.Context, r *usermodel.Role) error
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	CreatePermission(ctx context.Context, p *usermodel.Permission) error
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
