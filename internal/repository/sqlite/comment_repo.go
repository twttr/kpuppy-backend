package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/twttr/kpuppy-backend/internal/domain"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.content_id, c.user_id, c.text, c.spoiler, c.parent_id,
		       c.edited_at, c.deleted_at, c.created_at,
		       u.id, u.user_hash, u.display_name, u.avatar, u.is_banned, u.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`, id)

	return r.scanCommentWithUser(row)
}

func (r *CommentRepository) GetByContentID(ctx context.Context, contentID string) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.content_id, c.user_id, c.text, c.spoiler, c.parent_id,
		       c.edited_at, c.deleted_at, c.created_at,
		       u.id, u.user_hash, u.display_name, u.avatar, u.is_banned, u.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.content_id = ? AND c.parent_id IS NULL
		ORDER BY c.created_at DESC
	`, contentID)
	if err != nil {
		return nil, err
	}

	var comments []domain.Comment
	for rows.Next() {
		comment, err := r.scanCommentWithUserFromRows(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		comments = append(comments, *comment)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range comments {
		replies, err := r.GetReplies(ctx, comments[i].ID)
		if err != nil {
			return nil, err
		}
		comments[i].Replies = replies
	}

	return comments, nil
}

func (r *CommentRepository) GetReplies(ctx context.Context, parentID string) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.content_id, c.user_id, c.text, c.spoiler, c.parent_id,
		       c.edited_at, c.deleted_at, c.created_at,
		       u.id, u.user_hash, u.display_name, u.avatar, u.is_banned, u.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = ?
		ORDER BY c.created_at ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []domain.Comment
	for rows.Next() {
		reply, err := r.scanCommentWithUserFromRows(rows)
		if err != nil {
			return nil, err
		}
		replies = append(replies, *reply)
	}

	for i := range replies {
		nestedReplies, err := r.GetReplies(ctx, replies[i].ID)
		if err != nil {
			return nil, err
		}
		replies[i].Replies = nestedReplies
	}

	return replies, rows.Err()
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO comments (id, content_id, user_id, text, spoiler, parent_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, comment.ID, comment.ContentID, comment.UserID, comment.Text,
		boolToInt(comment.Spoiler), comment.ParentID, comment.CreatedAt.Unix())

	return err
}

func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	var editedAt *int64
	if comment.EditedAt != nil {
		ts := comment.EditedAt.Unix()
		editedAt = &ts
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE comments SET text = ?, spoiler = ?, edited_at = ? WHERE id = ?
	`, comment.Text, boolToInt(comment.Spoiler), editedAt, comment.ID)

	return err
}

func (r *CommentRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx, `
		UPDATE comments SET deleted_at = ? WHERE id = ?
	`, now, id)

	return err
}

func (r *CommentRepository) Restore(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE comments SET deleted_at = NULL WHERE id = ?
	`, id)

	return err
}

func (r *CommentRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]domain.Comment, int, error) {
	var countQuery, selectQuery string
	if includeDeleted {
		countQuery = `SELECT COUNT(*) FROM comments`
		selectQuery = `
			SELECT c.id, c.content_id, c.user_id, c.text, c.spoiler, c.parent_id,
			       c.edited_at, c.deleted_at, c.created_at,
			       u.id, u.user_hash, u.display_name, u.avatar, u.is_banned, u.created_at,
			       ct.kinopub_item_id
			FROM comments c
			JOIN users u ON c.user_id = u.id
			JOIN content ct ON c.content_id = ct.id
			ORDER BY c.created_at DESC
			LIMIT ? OFFSET ?
		`
	} else {
		countQuery = `SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL`
		selectQuery = `
			SELECT c.id, c.content_id, c.user_id, c.text, c.spoiler, c.parent_id,
			       c.edited_at, c.deleted_at, c.created_at,
			       u.id, u.user_hash, u.display_name, u.avatar, u.is_banned, u.created_at,
			       ct.kinopub_item_id
			FROM comments c
			JOIN users u ON c.user_id = u.id
			JOIN content ct ON c.content_id = ct.id
			WHERE c.deleted_at IS NULL
			ORDER BY c.created_at DESC
			LIMIT ? OFFSET ?
		`
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, selectQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var comment domain.Comment
		var user domain.User
		var parentID sql.NullString
		var editedAt, deletedAt sql.NullInt64
		var createdAt int64
		var avatar sql.NullString
		var isBanned int
		var userCreatedAt int64

		err := rows.Scan(
			&comment.ID, &comment.ContentID, &comment.UserID, &comment.Text,
			&comment.Spoiler, &parentID, &editedAt, &deletedAt, &createdAt,
			&user.ID, &user.UserHash, &user.DisplayName, &avatar, &isBanned, &userCreatedAt,
			&comment.KinopubItemID,
		)
		if err != nil {
			return nil, 0, err
		}

		comment.CreatedAt = time.Unix(createdAt, 0)
		if parentID.Valid {
			comment.ParentID = &parentID.String
		}
		if editedAt.Valid {
			t := time.Unix(editedAt.Int64, 0)
			comment.EditedAt = &t
		}
		if deletedAt.Valid {
			t := time.Unix(deletedAt.Int64, 0)
			comment.DeletedAt = &t
		}

		user.Avatar = nil
		if avatar.Valid {
			user.Avatar = &avatar.String
		}
		user.IsBanned = isBanned == 1
		user.CreatedAt = time.Unix(userCreatedAt, 0)
		comment.User = &user

		comments = append(comments, comment)
	}

	return comments, total, rows.Err()
}

func (r *CommentRepository) scanCommentWithUser(row *sql.Row) (*domain.Comment, error) {
	var comment domain.Comment
	var user domain.User
	var parentID sql.NullString
	var editedAt, deletedAt sql.NullInt64
	var createdAt int64
	var avatar sql.NullString
	var isBanned int
	var userCreatedAt int64

	err := row.Scan(
		&comment.ID, &comment.ContentID, &comment.UserID, &comment.Text, &comment.Spoiler, &parentID,
		&editedAt, &deletedAt, &createdAt,
		&user.ID, &user.UserHash, &user.DisplayName, &avatar, &isBanned, &userCreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		comment.ParentID = &parentID.String
	}
	if editedAt.Valid {
		t := time.Unix(editedAt.Int64, 0)
		comment.EditedAt = &t
	}
	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		comment.DeletedAt = &t
	}
	comment.CreatedAt = time.Unix(createdAt, 0)

	if avatar.Valid {
		user.Avatar = &avatar.String
	}
	user.IsBanned = isBanned == 1
	user.CreatedAt = time.Unix(userCreatedAt, 0)
	comment.User = &user

	return &comment, nil
}

func (r *CommentRepository) scanCommentWithUserFromRows(rows *sql.Rows) (*domain.Comment, error) {
	var comment domain.Comment
	var user domain.User
	var parentID sql.NullString
	var editedAt, deletedAt sql.NullInt64
	var createdAt int64
	var avatar sql.NullString
	var isBanned int
	var userCreatedAt int64

	err := rows.Scan(
		&comment.ID, &comment.ContentID, &comment.UserID, &comment.Text, &comment.Spoiler, &parentID,
		&editedAt, &deletedAt, &createdAt,
		&user.ID, &user.UserHash, &user.DisplayName, &avatar, &isBanned, &userCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		comment.ParentID = &parentID.String
	}
	if editedAt.Valid {
		t := time.Unix(editedAt.Int64, 0)
		comment.EditedAt = &t
	}
	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		comment.DeletedAt = &t
	}
	comment.CreatedAt = time.Unix(createdAt, 0)

	if avatar.Valid {
		user.Avatar = &avatar.String
	}
	user.IsBanned = isBanned == 1
	user.CreatedAt = time.Unix(userCreatedAt, 0)
	comment.User = &user

	return &comment, nil
}
