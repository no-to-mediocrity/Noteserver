package api

import (
	"encoding/json"
	"net/http"
	"noteserver/internal/pkg/actions"
	l "noteserver/internal/pkg/logger"
	"noteserver/internal/pkg/models"
	"noteserver/internal/pkg/notes"
	"noteserver/internal/pkg/responses"
	"noteserver/internal/pkg/yandex"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
)

func HandleLogin(w http.ResponseWriter, r *http.Request, db *pgx.Conn, jwtSecret []byte) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	storedUser, err := GetUserByUsername(user.Username, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "No such user", http.StatusInternalServerError)
		return
	}

	if storedUser == nil || !ComparePasswords(storedUser.Password, user.Password) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateJWTToken(storedUser, jwtSecret)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request, db *pgx.Conn, jwtSecret []byte) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user, err := GetUserFromToken(token, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = DeleteUser(*user, db)
	var response struct {
		Message string `json:"message"`
	}
	if err == nil {
		response.Message = "User deleted successfully"
	} else {
		response.Message = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleRegister(w http.ResponseWriter, r *http.Request, db *pgx.Conn) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	existingUser, err := GetUserByUsername(user.Username, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user.Password = hashedPassword

	err = SaveUser(user, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := struct {
		Message string `json:"message"`
	}{
		Message: "User registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleNotesAction(w http.ResponseWriter, r *http.Request, db *pgx.Conn, jwtSecret []byte, action actions.Type, apiTimeout int) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user, err := GetUserFromToken(token, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var note models.Note
	err = json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	switch action {
	case 1:
		CreateNoteHandler(w, r, db, user, &note, apiTimeout)
		return
	case 2:
		ReadNoteHandler(w, r, db, user, &note)
		return
	case 3:
		UpdateNoteHandler(w, r, db, user, &note, apiTimeout)
		return
	case 4:
		DeleteNoteHandler(w, r, db, user, &note)
		return
	default:
		return
	}
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, user *models.User, note *models.Note, apiTimeout int) {
	note_id, err := notes.CreateNote(db, note, user)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	note_id_string := strconv.Itoa(note_id)

	response := responses.CreateUpdateNote{}
	if err == nil {
		response.Status = "success"
		response.Message = "Note has been created successfully"
		response.NoteID = note_id_string
		spellcheck, err := yandex.Spellcheck(note.Content, apiTimeout)
		if err == nil {
			if len(spellcheck) == 0 {
				response.Spelling = "correct"
			} else {
				response.Spelling = "suggestions"
				response.SpellingSuggestions = &spellcheck
			}
		} else {
			response.Spelling = err.Error()
		}
	} else {
		response.SetError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ReadNoteHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, user *models.User, note *models.Note) {
	readnote, err := notes.ReadNote(db, note, user)
	if err != nil && err.Error() != "No matching notes found" {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := responses.ReadNote{}
	if err == nil {
		response.Status = "success"
		response.Message = "Note has been read successfully"
		response.Note = &readnote
	} else {
		response.SetError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, user *models.User, note *models.Note, apiTimeout int) {
	err := notes.UpdateNote(db, note, user)
	if err != nil && err.Error() != "No matching notes found" {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := responses.CreateUpdateNote{}
	if err == nil {
		response.Status = "success"
		response.Message = "Note has been updated successfully"
		response.NoteID = strconv.Itoa(note.ID)
		spellcheck, err := yandex.Spellcheck(note.Content, apiTimeout)
		if err == nil {
			if len(spellcheck) == 0 {
				response.Spelling = "correct"
			} else {
				response.Spelling = "suggestions"
				response.SpellingSuggestions = &spellcheck
			}
		} else {
			response.Spelling = err.Error()
		}
	} else {
		response.SetError(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, user *models.User, note *models.Note) {
	err := notes.DeleteNote(db, note, user)
	response := responses.DeleteNote{}
	if err == nil {
		response.Status = "success"
		response.Message = "Note has been deleted successfully"
	} else {
		response.SetError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleMultipleNotesAction(w http.ResponseWriter, r *http.Request, db *pgx.Conn, jwtSecret []byte) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user, err := GetUserFromToken(token, db)
	if err != nil {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	GetAllNotesHandler(w, r, db, user)
}

func GetAllNotesHandler(w http.ResponseWriter, r *http.Request, db *pgx.Conn, user *models.User) {
	notes, err := notes.GetAllNotes(db, user)
	if err != nil && err.Error() != "No notes found for the user" {
		l.Logger.Error("Error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := responses.AllNotes{}
	response.Status = "success"
	if len(notes) > 0 {
		response.Message = "All notes retrieved successfully"
		response.Notes = notes
	} else {
		response.Message = "No notes found"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
