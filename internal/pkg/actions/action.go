package actions

type Type int

const (
	CreateNote Type = iota + 1
	ReadNote
	UpdateNote
	DeleteNote
)
