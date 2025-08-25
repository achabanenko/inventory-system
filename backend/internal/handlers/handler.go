package handlers

import (
	"database/sql"
	"inventory/internal/config"
)

type Handler struct {
	DB     *sql.DB
	Config *config.Config
}

func New(db *sql.DB, cfg *config.Config) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
	}
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type PaginationParams struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Sort     string `query:"sort"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	Total      int64       `json:"total"`
}