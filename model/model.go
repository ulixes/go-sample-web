package model

import "time"

type Post struct {
	ID    int64
	Title string
	Text  string
	Time  time.Time
}

func NewPost(id int64, title string, text string) *Post {
	return &Post{id, title, text, time.Now()}
}

func LocalPost(title string, text string) *Post {
	return NewPost(-1, title, text)
}
