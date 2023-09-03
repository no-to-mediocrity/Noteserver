package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"noteserver/internal/pkg/actions"
	"noteserver/internal/pkg/api"
	l "noteserver/internal/pkg/logger"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v4"
)

func main() {
	var (
		port            string
		jwtSecretString string
		sqlServer       string
		apiTimeout      int
	)

	flag.StringVar(&port, "port", "8080", "Server port number")
	flag.StringVar(&jwtSecretString, "jwt-secret", "your-secret-key", "JWT secret key")
	flag.StringVar(&sqlServer, "sql-server", "postgresql://postgres:mysecretpassword@localhost:5432/postgres", "Parameters of SQL-Server")
	flag.IntVar(&apiTimeout, "timeout", 5, "External API timeout in seconds")

	flag.Parse()
	jwtSecret := []byte(jwtSecretString)

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Incorrect port number:", err)
	}
	if !isValidPort(portInt) {
		log.Fatal("Incorrect port number")
	}

	if !checkPostgreSQL(sqlServer) {
		log.Fatal("Incorrect SQL-server parameters")
	}
	err = l.InitLogger()
	if err != nil {
		log.Fatal("Failed to initialize", err)
	}
	connConfig, err := pgx.ParseConfig(sqlServer)
	if err != nil {
		l.Logger.Fatal("Failed to parse database URL:", err)
	}

	db, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		l.Logger.Fatal("Failed to connect to database:", err)
	}

	defer db.Close(context.Background())
	router := mux.NewRouter()
	router.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		api.HandleLogin(w, r, db, jwtSecret)
	}).Methods("POST")

	router.HandleFunc("/v1/register", func(w http.ResponseWriter, r *http.Request) {
		api.HandleRegister(w, r, db)
	}).Methods("POST")

	RegisterNoteRoutes(router, db, jwtSecret, apiTimeout)

	router.HandleFunc("/v1/allnotes", api.AuthenticateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		api.HandleMultipleNotesAction(w, r, db, jwtSecret)
	}, jwtSecret)).Methods("GET")

	router.HandleFunc("/v1/deleteuser", api.AuthenticateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		api.HandleDeleteUser(w, r, db, jwtSecret)
	}, jwtSecret)).Methods("DELETE")
	port = ":" + port
	l.Logger.Info("Server started on", port)
	fmt.Printf("Server started on %v\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func RegisterNoteRoutes(router *mux.Router, db *pgx.Conn, jwtSecret []byte, apiTimeout int) {
	actions_map := map[string]actions.Type{
		"POST":   actions.CreateNote,
		"GET":    actions.ReadNote,
		"PATCH":  actions.UpdateNote,
		"DELETE": actions.DeleteNote,
	}

	for method, action := range actions_map {
		RegisterNoteRoute(router, method, db, jwtSecret, action, apiTimeout)
	}
}

func RegisterNoteRoute(router *mux.Router, method string, db *pgx.Conn, jwtSecret []byte, action actions.Type, apiTimeout int) {
	router.HandleFunc("/v1/note", api.AuthenticateMiddleware(func(w http.ResponseWriter, r *http.Request) {
		api.HandleNotesAction(w, r, db, jwtSecret, action, apiTimeout)
	}, jwtSecret)).Methods(method)
}

func isValidPort(port int) bool {
	return port > 0 && port <= 65535
}

func checkPostgreSQL(inputString string) bool {
	pattern := `^postgresql://([^:]+):([^@]+)@([^:]+):(\d+)/([^/]+)$`
	match, _ := regexp.MatchString(pattern, inputString)

	return match
}
