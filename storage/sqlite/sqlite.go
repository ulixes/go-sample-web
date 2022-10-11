package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go-sample-web/model"
	"log"
)

type Storage struct {
	db *sql.DB
}

const (
	createTableSql = `CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, text TEXT, time DATETIME)`
	selectSql      = `SELECT id, title, text, time FROM posts ORDER BY time DESC`
	getByIdSql     = `SELECT id, title, text, time FROM posts where id=?`
	insertSql      = `INSERT INTO posts (title, text, time) VALUES (?, ?, ?)`
	updateSql      = `UPDATE posts SET title=?, text=?, time=? WHERE id = ?`
	existsSql      = `SELECT COUNT(*) FROM posts WHERE id = ?`
	deleteSql      = `DELETE FROM posts WHERE id = ?`
)

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, p *model.Post) error {
	if _, err := s.db.ExecContext(ctx, updateSql, p.Title, p.Text, p.Time, p.ID); err != nil {
		return fmt.Errorf("can't save post: %w", err)
	}
	return nil
}

func (s *Storage) Add(ctx context.Context, p *model.Post) error {
	result, err := s.db.ExecContext(ctx, insertSql, p.Title, p.Text, p.Time)
	if err != nil {
		return fmt.Errorf("can't save post: %w", err)
	}
	if p.ID, err = result.LastInsertId(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Exists(ctx context.Context, ID int) (bool, error) {
	rows, err := s.db.QueryContext(ctx, existsSql, ID)
	if err != nil {
		return false, fmt.Errorf("can't save post: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var count int
		if err = rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, errors.New("query unexpectedly didn't return any rows")
}

func (s *Storage) Delete(ctx context.Context, ID int) error {
	_, err := s.db.ExecContext(ctx, deleteSql, ID)
	if err != nil {
		return fmt.Errorf("can't delete post: %w", err)
	}
	return nil
}

func (s *Storage) GetAll(ctx context.Context) ([]*model.Post, error) {
	rows, err := s.db.QueryContext(ctx, selectSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*model.Post
	for rows.Next() {
		post := model.Post{}
		if err = rows.Scan(&post.ID, &post.Title, &post.Text, &post.Time); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return posts, nil
}

func (s *Storage) Find(ctx context.Context, ID int) (*model.Post, error) {
	rows, err := s.db.QueryContext(ctx, getByIdSql, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		post := model.Post{}
		if err = rows.Scan(&post.ID, &post.Title, &post.Text, &post.Time); err != nil {
			log.Fatal(err)
		}
		return &post, nil
	}
	return nil, fmt.Errorf("record not found by id: %d", ID)
}

func (s *Storage) Init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, createTableSql)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}
	return nil
}
