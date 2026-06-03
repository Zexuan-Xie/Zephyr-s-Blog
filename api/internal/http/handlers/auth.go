package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"xlab-blog/api/internal/auth"
	httpapi "xlab-blog/api/internal/http"
	"xlab-blog/api/internal/http/middleware"
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
		httpapi.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.service.Register(r.Context(), request.Email, request.Password, request.DisplayName)
	if err != nil {
		if errors.Is(err, users.ErrEmailExists) {
			httpapi.WriteError(w, http.StatusConflict, "email already exists")
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	result, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredential) {
			httpapi.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, user)
}
