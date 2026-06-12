package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"xlab-blog/api/internal/auth"
	"xlab-blog/api/internal/http/middleware"
	"xlab-blog/api/internal/http/respond"
	"xlab-blog/api/internal/users"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.service.Register(r.Context(), request.Email, request.Password, request.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrEmailExists):
			respond.Error(w, http.StatusConflict, "email already exists")
			return
		case strings.HasPrefix(err.Error(), "invalid email:"):
			respond.Error(w, http.StatusBadRequest, "invalid email")
			return
		case err.Error() == "password must be at least 8 characters":
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		default:
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}
	respond.JSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredential) {
			respond.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	respond.JSON(w, http.StatusOK, user)
}
