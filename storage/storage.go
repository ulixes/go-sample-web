package storage

import (
	"context"
	"go-sample-web/model"
)

type Storage interface {
	// Exists - return true if post exists
	Exists(ctx context.Context, ID int) (bool, error)
	// Delete - deletes posts from db
	Delete(ctx context.Context, ID int) error
	// Add - adds new post to db
	Add(ctx context.Context, p *model.Post) error
	// Save - saves existing post to db
	Save(ctx context.Context, p *model.Post) error
	// GetAll - returns all posts from db
	GetAll(ctx context.Context) ([]*model.Post, error)
	// Find - return post by id from db
	Find(ctx context.Context, ID int) (*model.Post, error)
}
