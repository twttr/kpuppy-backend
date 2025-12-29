package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

type ContentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kinopub_item_id, created_at
		FROM content WHERE id = ?
	`, id)

	return r.scanContent(row)
}

func (r *ContentRepository) GetByKinopubItemID(ctx context.Context, kinopubItemID int64) (*domain.Content, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kinopub_item_id, created_at
		FROM content WHERE kinopub_item_id = ?
	`, kinopubItemID)

	return r.scanContent(row)
}

func (r *ContentRepository) Create(ctx context.Context, content *domain.Content) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO content (id, kinopub_item_id, created_at)
		VALUES (?, ?, ?)
	`, content.ID, content.KinopubItemID, content.CreatedAt.Unix())

	return err
}

func (r *ContentRepository) GetOrCreate(ctx context.Context, kinopubItemID int64) (*domain.Content, error) {
	content, err := r.GetByKinopubItemID(ctx, kinopubItemID)
	if err == nil {
		return content, nil
	}
	if err != domain.ErrContentNotFound {
		return nil, err
	}

	content = &domain.Content{
		ID:            uuid.New().String(),
		KinopubItemID: kinopubItemID,
		CreatedAt:     time.Now(),
	}

	if err := r.Create(ctx, content); err != nil {
		existing, getErr := r.GetByKinopubItemID(ctx, kinopubItemID)
		if getErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return content, nil
}

func (r *ContentRepository) scanContent(row *sql.Row) (*domain.Content, error) {
	var content domain.Content
	var createdAt int64

	err := row.Scan(&content.ID, &content.KinopubItemID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrContentNotFound
	}
	if err != nil {
		return nil, err
	}

	content.CreatedAt = time.Unix(createdAt, 0)

	return &content, nil
}
