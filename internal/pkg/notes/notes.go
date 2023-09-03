package notes

import (
	"context"
	"fmt"
	"noteserver/internal/pkg/models"
	"time"

	"github.com/jackc/pgx/v4"
)

func ReadNote(conn *pgx.Conn, note *models.Note, user *models.User) (models.Note, error) {
	var readnote models.Note
	err := conn.QueryRow(
		context.Background(),
		"SELECT * FROM Notes WHERE note_id = $1 AND user_id = $2",
		note.ID, user.ID,
	).Scan(&readnote.ID, &readnote.UserID, &readnote.Title, &readnote.Content, &readnote.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Note{}, fmt.Errorf("No matching notes found")
		}
		return models.Note{}, err
	}

	return readnote, nil
}

func DeleteNote(conn *pgx.Conn, note *models.Note, user *models.User) error {
	result, err := conn.Exec(
		context.Background(),
		"DELETE FROM Notes WHERE note_id = $1 AND user_id = $2",
		note.ID, user.ID,
	)
	if result.RowsAffected() == 0 {
		return fmt.Errorf("No matching notes found")
	}
	return err
}

func UpdateNote(conn *pgx.Conn, note *models.Note, user *models.User) error {
	existingNote, err := ReadNote(conn, note, user)
	if err != nil {
		return err
	}
	result, err := conn.Exec(
		context.Background(),
		"UPDATE Notes SET title = $1, content = $2 WHERE note_id = $3 AND user_id = $4",
		note.Title, note.Content, existingNote.ID, user.ID,
	)
	if result.RowsAffected() == 0 {
		return fmt.Errorf("No matching notes found")
	}
	if err != nil {
		return err
	}

	return nil
}

func CreateNote(conn *pgx.Conn, note *models.Note, user *models.User) (int, error) {
	var noteID int
	err := conn.QueryRow(context.Background(),
		"INSERT INTO Notes(user_id, title, content, created_at) VALUES($1, $2, $3, $4) RETURNING note_id",
		user.ID, note.Title, note.Content, time.Now()).Scan(&noteID)
	if err != nil {
		return 0, err
	}
	return noteID, nil
}

func GetAllNotes(conn *pgx.Conn, user *models.User) ([]models.Note, error) {
	rows, err := conn.Query(
		context.Background(),
		"SELECT * FROM Notes WHERE user_id = $1",
		user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}

func DeleteAllNotes(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "DELETE FROM Notes")
	return err
}
