package responses

import "noteserver/internal/pkg/models"

type AllNotes struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Notes   []models.Note `json:"notes"`
}
