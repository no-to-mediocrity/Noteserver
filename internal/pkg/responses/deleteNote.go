package responses

type DeleteNote struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *DeleteNote) SetError(message string) {
	c.Status = "error"
	c.Message = message
}
