package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"xlab-blog/api/internal/http/respond"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if h.pool == nil {
		respond.JSON(w, http.StatusOK, HealthResponse{Status: "ok"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.pool.Ping(ctx); err != nil {
		respond.JSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "degraded", Database: "unavailable"})
		return
	}
	respond.JSON(w, http.StatusOK, HealthResponse{Status: "ok", Database: "ok"})
}
