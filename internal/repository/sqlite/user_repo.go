package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/twttr/kpuppy-backend/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_hash, display_name, avatar, is_banned, created_at
		FROM users WHERE id = ?
	`, id)

	return r.scanUser(row)
}

func (r *UserRepository) GetByUserHash(ctx context.Context, userHash string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_hash, display_name, avatar, is_banned, created_at
		FROM users WHERE user_hash = ?
	`, userHash)

	return r.scanUser(row)
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, user_hash, display_name, avatar, is_banned, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.UserHash, user.DisplayName, user.Avatar, boolToInt(user.IsBanned), user.CreatedAt.Unix())

	return err
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, id string, avatar *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET avatar = ? WHERE id = ?
	`, avatar, id)

	return err
}

func (r *UserRepository) SetBanned(ctx context.Context, id string, banned bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET is_banned = ? WHERE id = ?
	`, boolToInt(banned), id)

	return err
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_hash, display_name, avatar, is_banned, created_at
		FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, *user)
	}

	return users, total, rows.Err()
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	var avatar sql.NullString
	var isBanned int
	var createdAt int64

	err := row.Scan(&user.ID, &user.UserHash, &user.DisplayName, &avatar, &isBanned, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if avatar.Valid {
		user.Avatar = &avatar.String
	}
	user.IsBanned = isBanned == 1
	user.CreatedAt = time.Unix(createdAt, 0)

	return &user, nil
}

func (r *UserRepository) scanUserFromRows(rows *sql.Rows) (*domain.User, error) {
	var user domain.User
	var avatar sql.NullString
	var isBanned int
	var createdAt int64

	err := rows.Scan(&user.ID, &user.UserHash, &user.DisplayName, &avatar, &isBanned, &createdAt)
	if err != nil {
		return nil, err
	}

	if avatar.Valid {
		user.Avatar = &avatar.String
	}
	user.IsBanned = isBanned == 1
	user.CreatedAt = time.Unix(createdAt, 0)

	return &user, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
