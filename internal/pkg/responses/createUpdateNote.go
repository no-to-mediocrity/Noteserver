package responses

import "noteserver/internal/pkg/models"

type CreateUpdateNote struct {
	Status              string                   `json:"status"`
	Message             string                   `json:"message"`
	NoteID              string                   `json:"note_id"`
	Spelling            string                   `json:"spelling"`
	SpellingSuggestions *[]models.SpellcheckData `json:"spelling_suggestion"`
}

func (c *CreateUpdateNote) SetError(message string) {
	c.Status = "error"
	c.Message = message
}
