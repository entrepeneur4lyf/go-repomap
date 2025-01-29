package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/web/models"
	"example.com/web/services"
)

type UserHandler struct {
	userService *services.UserService
}

func HandleUsers() http.HandlerFunc {
	handler := &UserHandler{
		userService: services.NewUserService(),
	}
	return handler.handle
}

func (h *UserHandler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users := h.userService.ListUsers()
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdUser := h.userService.CreateUser(user)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}
