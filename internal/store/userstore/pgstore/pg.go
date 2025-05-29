// Package pgstore implements the userstore.Store interface using PostgreSQL
package pgstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/usermodel"
)

type PGStore struct {
	db *sql.DB
}

func New(db *sql.DB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *PGStore) GetUserByID(ctx context.Context, ID string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, first_name, last_name, email, password_hash, mfa_secret FROM users WHERE id = $1
	`, ID)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) GetUserByUsername(ctx context.Context, username string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, first_name, last_name, email, password_hash, mfa_secret FROM users WHERE username = $1
	`, username)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, mfa_secret FROM users WHERE email = $1
	`, email)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) GetUserWithPermissions(ctx context.Context, userID string) (*usermodel.User, error) {
	u := &usermodel.User{ID: userID, Roles: []usermodel.Role{}}

	err := s.db.QueryRowContext(ctx, `
		SELECT username, first_name, last_name, email, password_hash, mfa_secret FROM users WHERE id = $1
	`, userID).Scan(&u.Username, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.TOTPSecret)
	if err != nil {
		return nil, err
	}

	roleRows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	for roleRows.Next() {
		var role usermodel.Role
		if err := roleRows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}

		permRows, err := s.db.QueryContext(ctx, `
			SELECT p.id, p.name, p.description
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = $1
		`, role.ID)
		if err != nil {
			return nil, err
		}
		defer permRows.Close()
		for permRows.Next() {
			var perm usermodel.Permission
			if err := permRows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
				permRows.Close()
				return nil, err
			}
			role.Permissions = append(role.Permissions, perm)
		}
		permRows.Close()

		u.Roles = append(u.Roles, role)

	}

	scopeRows, err := s.db.QueryContext(ctx, `
		SELECT resource, scope_value
		FROM user_scopes
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, err
	}
	defer scopeRows.Close()
	scopes := make(map[string][]string)
	for scopeRows.Next() {
		var resource, value string
		if err := scopeRows.Scan(&resource, &value); err != nil {
			return nil, err
		}
		scopes[resource] = append(scopes[resource], value)
	}

	u.Scopes = scopes
	return u, nil
}

func (s *PGStore) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	return nil
}

func (s *PGStore) GetUserBySSO(ctx context.Context, provider, ssoID string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, mfa_secret, sso_provider, sso_id, last_login
		FROM users
		WHERE sso_provider = $1 AND sso_id = $2
	`, provider, ssoID)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TOTPSecret, &u.SSOProvider, &u.SSOID, &u.LastLogin); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) SaveUser(ctx context.Context, u *usermodel.User) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET
			last_login = $1,
			sso_provider = $2,
			sso_id = $3
		WHERE id = $4
	`, u.LastLogin, u.SSOProvider, u.SSOID, u.ID)

	return err
}

func (s *PGStore) CreateRole(ctx context.Context, r *usermodel.Role) error { return nil }

func (s *PGStore) CreatePermission(ctx context.Context, p *usermodel.Permission) error { return nil }
func (s *PGStore) AttachPermissionToRole(ctx context.Context, roleID, permID string) error {
	return nil
}

// Profile management methods

func (s *PGStore) GetUserProfile(ctx context.Context, userID string) (*usermodel.UserProfile, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT user_id, full_name, phone, avatar_url, updated_at
		FROM user_profiles
		WHERE user_id = $1
	`, userID)

	profile := &usermodel.UserProfile{}
	err := row.Scan(&profile.UserID, &profile.FullName, &profile.Phone, &profile.AvatarURL, &profile.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty profile if none exists
			return &usermodel.UserProfile{UserID: userID}, nil
		}
		return nil, err
	}
	return profile, nil
}

func (s *PGStore) CreateUserProfile(ctx context.Context, profile *usermodel.UserProfile) error {
	query := `
		INSERT INTO user_profiles (user_id, full_name, phone, avatar_url, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			full_name = EXCLUDED.full_name,
			phone = EXCLUDED.phone,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query, profile.UserID, profile.FullName, profile.Phone, profile.AvatarURL)
	return err
}

func (s *PGStore) UpdateUserProfile(ctx context.Context, profile *usermodel.UserProfile) error {
	profile.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE user_profiles
		SET full_name = $2, phone = $3, avatar_url = $4, updated_at = $5
		WHERE user_id = $1
	`, profile.UserID, profile.FullName, profile.Phone, profile.AvatarURL, profile.UpdatedAt)
	return err
}

// Settings management methods

func (s *PGStore) GetUserSettings(ctx context.Context, userID string) (usermodel.UserSettings, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT key, value
		FROM user_settings
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(usermodel.UserSettings)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, nil
}

func (s *PGStore) GetUserSetting(ctx context.Context, userID, key string) (*usermodel.UserSetting, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT user_id, key, value
		FROM user_settings
		WHERE user_id = $1 AND key = $2
	`, userID, key)

	setting := &usermodel.UserSetting{}
	err := row.Scan(&setting.UserID, &setting.Key, &setting.Value)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *PGStore) SetUserSetting(ctx context.Context, userID, key string, value []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_settings (user_id, key, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, key) DO UPDATE SET
			value = EXCLUDED.value
	`, userID, key, value)
	return err
}

func (s *PGStore) DeleteUserSetting(ctx context.Context, userID, key string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM user_settings
		WHERE user_id = $1 AND key = $2
	`, userID, key)
	return err
}

// Password management

func (s *PGStore) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1
	`, userID, passwordHash)
	return err
}

// Complete user operations

func (s *PGStore) GetCompleteUser(ctx context.Context, userID string) (*usermodel.CompleteUser, error) {
	// Get the base user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get the user profile
	profile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get the user settings
	settings, err := s.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return &usermodel.CompleteUser{
		User:     user,
		Profile:  profile,
		Settings: settings,
	}, nil
}

// GetAllUsersWithPermissions retrieves all users with their roles and permissions
func (s *PGStore) GetAllUsersWithPermissions(ctx context.Context) ([]*usermodel.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, username, first_name, last_name, email, created_at, updated_at, last_login
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*usermodel.User
	for rows.Next() {
		u := &usermodel.User{Roles: []usermodel.Role{}}
		err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.LastLogin)
		if err != nil {
			return nil, err
		}

		// Get roles for this user
		roleRows, err := s.db.QueryContext(ctx, `
			SELECT r.id, r.name, r.description
			FROM roles r
			JOIN user_roles ur ON ur.role_id = r.id
			WHERE ur.user_id = $1
		`, u.ID)
		if err != nil {
			return nil, err
		}

		for roleRows.Next() {
			var role usermodel.Role
			if err := roleRows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
				roleRows.Close()
				return nil, err
			}

			// Get permissions for this role
			permRows, err := s.db.QueryContext(ctx, `
				SELECT p.id, p.name, p.description
				FROM permissions p
				JOIN role_permissions rp ON rp.permission_id = p.id
				WHERE rp.role_id = $1
			`, role.ID)
			if err != nil {
				roleRows.Close()
				return nil, err
			}

			for permRows.Next() {
				var perm usermodel.Permission
				if err := permRows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
					permRows.Close()
					roleRows.Close()
					return nil, err
				}
				role.Permissions = append(role.Permissions, perm)
			}
			permRows.Close()
			u.Roles = append(u.Roles, role)
		}
		roleRows.Close()
		users = append(users, u)
	}
	return users, nil
}

// CreateUser creates a new user in the database
func (s *PGStore) CreateUser(ctx context.Context, u *usermodel.User) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, first_name, last_name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, u.Username, u.FirstName, u.LastName, u.Email, u.PasswordHash, u.CreatedAt, u.UpdatedAt).Scan(&u.ID)

	return err
}

// UpdateUser updates an existing user in the database
func (s *PGStore) UpdateUser(ctx context.Context, u *usermodel.User) error {
	u.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET username = $2, first_name = $3, last_name = $4, email = $5, updated_at = $6
		WHERE id = $1
	`, u.ID, u.Username, u.FirstName, u.LastName, u.Email, u.UpdatedAt)

	return err
}

// DeleteUser removes a user from the database
func (s *PGStore) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}
