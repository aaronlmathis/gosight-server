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

func (s *PGStore) GetUserBySSO(ctx context.Context, provider string, ssoID string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.first_name, u.last_name, u.email, u.password_hash, u.mfa_secret 
		FROM users u 
		JOIN user_sso_providers usp ON u.id = usp.user_id 
		WHERE usp.provider = $1 AND usp.sso_id = $2
	`, provider, ssoID)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
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

// Role management methods

func (s *PGStore) GetRoles(ctx context.Context) ([]*usermodel.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description 
		FROM roles 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*usermodel.Role
	for rows.Next() {
		role := &usermodel.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *PGStore) GetRole(ctx context.Context, roleID string) (*usermodel.Role, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, description 
		FROM roles 
		WHERE id = $1
	`, roleID)

	role := &usermodel.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}

	// Get permissions for this role
	permRows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.description
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer permRows.Close()

	for permRows.Next() {
		perm := usermodel.Permission{}
		err := permRows.Scan(&perm.ID, &perm.Name, &perm.Description)
		if err != nil {
			return nil, err
		}
		role.Permissions = append(role.Permissions, perm)
	}

	return role, nil
}

func (s *PGStore) CreateRole(ctx context.Context, r *usermodel.Role) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO roles (name, description)
		VALUES ($1, $2)
		RETURNING id
	`, r.Name, r.Description).Scan(&r.ID)

	return err
}

func (s *PGStore) UpdateRole(ctx context.Context, r *usermodel.Role) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE roles
		SET name = $2, description = $3
		WHERE id = $1
	`, r.ID, r.Name, r.Description)

	return err
}

func (s *PGStore) DeleteRole(ctx context.Context, roleID string) error {
	// First remove all role-permission associations
	_, err := s.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return err
	}

	// Then remove all user-role associations
	_, err = s.db.ExecContext(ctx, `DELETE FROM user_roles WHERE role_id = $1`, roleID)
	if err != nil {
		return err
	}

	// Finally, delete the role
	_, err = s.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, roleID)
	return err
}

func (s *PGStore) GetRolePermissions(ctx context.Context, roleID string) ([]*usermodel.Permission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.description
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*usermodel.Permission
	for rows.Next() {
		perm := &usermodel.Permission{}
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Description)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func (s *PGStore) AssignPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error {
	for _, permID := range permissionIDs {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			VALUES ($1, $2)
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`, roleID, permID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error {
	for _, permID := range permissionIDs {
		_, err := s.db.ExecContext(ctx, `
			DELETE FROM role_permissions
			WHERE role_id = $1 AND permission_id = $2
		`, roleID, permID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) GetUsersWithRole(ctx context.Context, roleID string) ([]*usermodel.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.first_name, u.last_name, u.email, u.created_at, u.updated_at, u.last_login
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		WHERE ur.role_id = $1
		ORDER BY u.username
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*usermodel.User
	for rows.Next() {
		user := &usermodel.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Permission management methods

func (s *PGStore) GetPermissions(ctx context.Context) ([]*usermodel.Permission, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description 
		FROM permissions 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*usermodel.Permission
	for rows.Next() {
		perm := &usermodel.Permission{}
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Description)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func (s *PGStore) GetPermission(ctx context.Context, permissionID string) (*usermodel.Permission, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, description 
		FROM permissions 
		WHERE id = $1
	`, permissionID)

	perm := &usermodel.Permission{}
	err := row.Scan(&perm.ID, &perm.Name, &perm.Description)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *PGStore) CreatePermission(ctx context.Context, p *usermodel.Permission) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO permissions (name, description)
		VALUES ($1, $2)
		RETURNING id
	`, p.Name, p.Description).Scan(&p.ID)

	return err
}

func (s *PGStore) UpdatePermission(ctx context.Context, p *usermodel.Permission) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE permissions
		SET name = $2, description = $3
		WHERE id = $1
	`, p.ID, p.Name, p.Description)

	return err
}

func (s *PGStore) DeletePermission(ctx context.Context, permissionID string) error {
	// First remove all role-permission associations
	_, err := s.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE permission_id = $1`, permissionID)
	if err != nil {
		return err
	}

	// Then delete the permission
	_, err = s.db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, permissionID)
	return err
}

func (s *PGStore) GetRolesWithPermission(ctx context.Context, permissionID string) ([]*usermodel.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN role_permissions rp ON rp.role_id = r.id
		WHERE rp.permission_id = $1
		ORDER BY r.name
	`, permissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*usermodel.Role
	for rows.Next() {
		role := &usermodel.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// User role management methods

func (s *PGStore) GetUserRoles(ctx context.Context, userID string) ([]*usermodel.Role, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*usermodel.Role
	for rows.Next() {
		role := &usermodel.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *PGStore) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID)
	return err
}

func (s *PGStore) AssignRolesToUser(ctx context.Context, userID string, roleIDs []string) error {
	for _, roleID := range roleIDs {
		err := s.AssignRoleToUser(ctx, userID, roleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) RemoveRolesFromUser(ctx context.Context, userID string, roleIDs []string) error {
	for _, roleID := range roleIDs {
		_, err := s.db.ExecContext(ctx, `
			DELETE FROM user_roles
			WHERE user_id = $1 AND role_id = $2
		`, userID, roleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) AttachPermissionToRole(ctx context.Context, roleID, permID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, roleID, permID)
	return err
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

// SaveUser creates or updates a user in the database
func (s *PGStore) SaveUser(ctx context.Context, u *usermodel.User) error {
	// Check if user exists
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, u.ID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return s.UpdateUser(ctx, u)
	} else {
		return s.CreateUser(ctx, u)
	}
}
