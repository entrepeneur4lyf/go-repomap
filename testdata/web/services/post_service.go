package services

import (
	"sync"

	"example.com/web/models"
)

type PostService struct {
	posts    map[string]models.Post
	comments map[string][]models.PostComment
	mu       sync.RWMutex
}

func NewPostService() *PostService {
	return &PostService{
		posts:    make(map[string]models.Post),
		comments: make(map[string][]models.PostComment),
	}
}

func (s *PostService) CreatePost(post models.Post) models.Post {
	s.mu.Lock()
	defer s.mu.Unlock()

	post.BeforeCreate()
	s.posts[post.ID] = post
	return post
}

func (s *PostService) GetPost(id string) (models.Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[id]
	return post, exists
}

func (s *PostService) UpdatePost(post models.Post) (models.Post, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[post.ID]; !exists {
		return models.Post{}, false
	}

	post.BeforeUpdate()
	s.posts[post.ID] = post
	return post, true
}

func (s *PostService) DeletePost(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[id]; !exists {
		return false
	}

	delete(s.posts, id)
	delete(s.comments, id)
	return true
}

func (s *PostService) ListPosts() []models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]models.Post, 0, len(s.posts))
	for _, post := range s.posts {
		posts = append(posts, post)
	}
	return posts
}

func (s *PostService) GetPostsByUser(userID string) []models.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userPosts []models.Post
	for _, post := range s.posts {
		if post.AuthorID == userID {
			userPosts = append(userPosts, post)
		}
	}
	return userPosts
}

func (s *PostService) AddComment(comment models.PostComment) models.PostComment {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment.BeforeCreate()
	s.comments[comment.PostID] = append(s.comments[comment.PostID], comment)
	return comment
}

func (s *PostService) GetComments(postID string) []models.PostComment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.comments[postID]
}
