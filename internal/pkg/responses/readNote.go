package responses

import "noteserver/internal/pkg/models"

type ReadNote struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Note    *models.Note `json:"note"`
}

func (c *ReadNote) SetError(message string) {
	c.Status = "error"
	c.Message = message
}
