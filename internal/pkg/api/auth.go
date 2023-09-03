package api

import (
	"context"
	"database/sql"
	"errors"
	"noteserver/internal/pkg/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

func GetUserByUsername(username string, db *pgx.Conn) (*models.User, error) {
	row := db.QueryRow(context.Background(), "SELECT user_id, username, password_hash FROM users WHERE username = $1", username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err == sql.ErrNoRows || err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByID(userID int, db *pgx.Conn) (*models.User, error) {
	row := db.QueryRow(context.Background(), "SELECT user_id, username, password_hash FROM users WHERE user_id = $1", userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err == sql.ErrNoRows || err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func ComparePasswords(hashedPwd string, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}

func GenerateJWTToken(user *models.User, jwtSecret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

func HashPassword(password string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPwd), nil
}

func SaveUser(user models.User, db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), "INSERT INTO users (username, password_hash) VALUES ($1, $2)", user.Username, user.Password)
	return err
}

func DeleteUser(user models.User, db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), "DELETE FROM notes WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	_, err = db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", user.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetUserFromToken(token *jwt.Token, db *pgx.Conn) (*models.User, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token claims")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("User ID not found in token claims")
	}
	user, err := GetUserByUsername(username, db)
	if err != nil {
		return nil, err
	}

	return user, nil
}
