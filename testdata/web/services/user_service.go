package services

import (
	"sync"

	"example.com/web/models"
)

type UserService struct {
	users map[string]models.User
	mu    sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]models.User),
	}
}

func (s *UserService) CreateUser(user models.User) models.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real application, we would:
	// 1. Generate a proper UUID
	// 2. Hash the password
	// 3. Store in a database
	// 4. Handle errors properly

	user.BeforeCreate()
	s.users[user.ID] = user
	return user
}

func (s *UserService) GetUser(id string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

func (s *UserService) UpdateUser(user models.User) (models.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return models.User{}, false
	}

	user.BeforeUpdate()
	s.users[user.ID] = user
	return user, true
}

func (s *UserService) DeleteUser(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return false
	}

	delete(s.users, id)
	return true
}

func (s *UserService) ListUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}
