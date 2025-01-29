package models

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

type Post struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	AuthorID  string     `json:"author_id"`
	Status    PostStatus `json:"status"`
	Tags      []string   `json:"tags"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PostComment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Post) Validate() bool {
	return p.Title != "" && p.Content != "" && p.AuthorID != ""
}

func (p *Post) BeforeCreate() {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = PostStatusDraft
	}
}

func (p *Post) BeforeUpdate() {
	p.UpdatedAt = time.Now()
}

func (pc *PostComment) Validate() bool {
	return pc.Content != "" && pc.AuthorID != "" && pc.PostID != ""
}

func (pc *PostComment) BeforeCreate() {
	now := time.Now()
	pc.CreatedAt = now
	pc.UpdatedAt = now
}

func (pc *PostComment) BeforeUpdate() {
	pc.UpdatedAt = time.Now()
}
