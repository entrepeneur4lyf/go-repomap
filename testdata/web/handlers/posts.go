package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/web/models"
	"example.com/web/services"
)

type PostHandler struct {
	postService *services.PostService
}

func HandlePosts() http.HandlerFunc {
	handler := &PostHandler{
		postService: services.NewPostService(),
	}
	return handler.handle
}

func (h *PostHandler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	var posts []models.Post
	if userID != "" {
		posts = h.postService.GetPostsByUser(userID)
	} else {
		posts = h.postService.ListPosts()
	}
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) create(w http.ResponseWriter, r *http.Request) {
	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdPost := h.postService.CreatePost(post)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost)
}
