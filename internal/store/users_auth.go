package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session represents an authenticated user web/API session.
type Session struct {
	ID        string
	UserID    string
	TokenHash string
	CSRFToken string
	IPAddress string
	UserAgent string
	CreatedAt string
	ExpiresAt string
}

// SSHKey represents an enrolled public key for Git SSH transport.
type SSHKey struct {
	ID          string
	UserID      string
	Name        string
	Fingerprint string
	PublicKey   string
	KeyType     string
	CreatedAt   string
}

// CreateUserWithPassword inserts a new user with an initial password hash, role, and status.
func (db *DB) CreateUserWithPassword(ctx context.Context, u User) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if u.CreatedAt == "" {
		u.CreatedAt = now
	}
	if u.Role == "" {
		u.Role = "member"
	}
	if u.Status == "" {
		u.Status = "active"
	}

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO users (id, email, login, name, password_hash, role, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, nullStr(u.Login), nullStr(u.Name), nullStr(u.PasswordHash), u.Role, u.Status, u.CreatedAt,
	)
	if err != nil {
		if isConflict(err) {
			return fmt.Errorf("%w: user with email %q already exists", ErrConflict, u.Email)
		}
		return err
	}
	return nil
}

// GetUserByEmail finds a user by exact email.
func (db *DB) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	var login, name, pass, role, status sql.NullString
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, email, login, name, password_hash, role, status, created_at
		FROM users WHERE email=?`, email).
		Scan(&u.ID, &u.Email, &login, &name, &pass, &role, &status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	u.Login = login.String
	u.Name = name.String
	u.PasswordHash = pass.String
	u.Role = role.String
	if u.Role == "" {
		u.Role = "member"
	}
	u.Status = status.String
	if u.Status == "" {
		u.Status = "active"
	}
	return u, err
}

// GetUserByLogin finds a user by login username.
func (db *DB) GetUserByLogin(ctx context.Context, login string) (User, error) {
	var u User
	var name, pass, role, status sql.NullString
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, email, login, name, password_hash, role, status, created_at
		FROM users WHERE login=?`, login).
		Scan(&u.ID, &u.Email, &u.Login, &name, &pass, &role, &status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	u.Name = name.String
	u.PasswordHash = pass.String
	u.Role = role.String
	if u.Role == "" {
		u.Role = "member"
	}
	u.Status = status.String
	if u.Status == "" {
		u.Status = "active"
	}
	return u, err
}

// ListUsers retrieves all registered users.
func (db *DB) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, email, login, name, password_hash, role, status, created_at
		FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var login, name, pass, role, status sql.NullString
		if err := rows.Scan(&u.ID, &u.Email, &login, &name, &pass, &role, &status, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Login = login.String
		u.Name = name.String
		u.PasswordHash = pass.String
		u.Role = role.String
		if u.Role == "" {
			u.Role = "member"
		}
		u.Status = status.String
		if u.Status == "" {
			u.Status = "active"
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserPassword updates the password hash for a user.
func (db *DB) UpdateUserPassword(ctx context.Context, userID, newHash string) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, newHash, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserStatus updates the active/suspended status for a user.
func (db *DB) UpdateUserStatus(ctx context.Context, userID, status string) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE users SET status=? WHERE id=?`, status, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserRole updates the enterprise role (admin, member, service) for a user.
func (db *DB) UpdateUserRole(ctx context.Context, userID, role string) error {
	res, err := db.sql.ExecContext(ctx, `UPDATE users SET role=? WHERE id=?`, role, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUser deletes a user and cascades deletions.
func (db *DB) DeleteUser(ctx context.Context, userID string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM users WHERE id=?`, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession stores a new user web session.
func (db *DB) CreateSession(ctx context.Context, s Session) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO user_sessions (id, user_id, token_hash, csrf_token, ip_address, user_agent, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.UserID, s.TokenHash, s.CSRFToken, nullStr(s.IPAddress), nullStr(s.UserAgent), s.CreatedAt, s.ExpiresAt,
	)
	if err != nil {
		if isConflict(err) {
			return fmt.Errorf("%w: session", ErrConflict)
		}
		return err
	}
	return nil
}

// GetSessionByHash retrieves an active session and associated user by session token hash.
func (db *DB) GetSessionByHash(ctx context.Context, hash string) (Session, User, error) {
	var s Session
	var u User
	var ip, ua, login, name, pass, role, status sql.NullString

	err := db.sql.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.token_hash, s.csrf_token, s.ip_address, s.user_agent, s.created_at, s.expires_at,
		       u.id, u.email, u.login, u.name, u.password_hash, u.role, u.status, u.created_at
		FROM user_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token_hash = ?`, hash).
		Scan(
			&s.ID, &s.UserID, &s.TokenHash, &s.CSRFToken, &ip, &ua, &s.CreatedAt, &s.ExpiresAt,
			&u.ID, &u.Email, &login, &name, &pass, &role, &status, &u.CreatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, User{}, ErrNotFound
	}
	if err != nil {
		return Session{}, User{}, err
	}

	s.IPAddress = ip.String
	s.UserAgent = ua.String
	u.Login = login.String
	u.Name = name.String
	u.PasswordHash = pass.String
	u.Role = role.String
	if u.Role == "" {
		u.Role = "member"
	}
	u.Status = status.String
	if u.Status == "" {
		u.Status = "active"
	}

	return s, u, nil
}

// DeleteSession invalidates a single session.
func (db *DB) DeleteSession(ctx context.Context, id string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM user_sessions WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSessionsForUser invalidates all sessions for a specific user (e.g. on password change).
func (db *DB) DeleteSessionsForUser(ctx context.Context, userID string) error {
	_, err := db.sql.ExecContext(ctx, `DELETE FROM user_sessions WHERE user_id=?`, userID)
	return err
}

// AddUserSSHKey stores an enrolled public key.
func (db *DB) AddUserSSHKey(ctx context.Context, k SSHKey) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if k.CreatedAt == "" {
		k.CreatedAt = now
	}
	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO user_ssh_keys (id, user_id, name, fingerprint, public_key, key_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		k.ID, k.UserID, k.Name, k.Fingerprint, k.PublicKey, k.KeyType, k.CreatedAt,
	)
	if err != nil {
		if isConflict(err) {
			return fmt.Errorf("%w: ssh key fingerprint already enrolled", ErrConflict)
		}
		return err
	}
	return nil
}

// ListUserSSHKeys lists all SSH public keys enrolled by a user.
func (db *DB) ListUserSSHKeys(ctx context.Context, userID string) ([]SSHKey, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, user_id, name, fingerprint, public_key, key_type, created_at
		FROM user_ssh_keys WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []SSHKey
	for rows.Next() {
		var k SSHKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.Fingerprint, &k.PublicKey, &k.KeyType, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetSSHKeyByFingerprint retrieves a key and its owning user by fingerprint (used by SSH server).
func (db *DB) GetSSHKeyByFingerprint(ctx context.Context, fingerprint string) (SSHKey, User, error) {
	var k SSHKey
	var u User
	var login, name, pass, role, status sql.NullString

	err := db.sql.QueryRowContext(ctx, `
		SELECT k.id, k.user_id, k.name, k.fingerprint, k.public_key, k.key_type, k.created_at,
		       u.id, u.email, u.login, u.name, u.password_hash, u.role, u.status, u.created_at
		FROM user_ssh_keys k
		JOIN users u ON k.user_id = u.id
		WHERE k.fingerprint = ?`, fingerprint).
		Scan(
			&k.ID, &k.UserID, &k.Name, &k.Fingerprint, &k.PublicKey, &k.KeyType, &k.CreatedAt,
			&u.ID, &u.Email, &login, &name, &pass, &role, &status, &u.CreatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return SSHKey{}, User{}, ErrNotFound
	}
	if err != nil {
		return SSHKey{}, User{}, err
	}

	u.Login = login.String
	u.Name = name.String
	u.PasswordHash = pass.String
	u.Role = role.String
	if u.Role == "" {
		u.Role = "member"
	}
	u.Status = status.String
	if u.Status == "" {
		u.Status = "active"
	}

	return k, u, nil
}

// DeleteUserSSHKey removes an enrolled SSH key.
func (db *DB) DeleteUserSSHKey(ctx context.Context, userID, keyID string) error {
	res, err := db.sql.ExecContext(ctx, `DELETE FROM user_ssh_keys WHERE id=? AND user_id=?`, keyID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
