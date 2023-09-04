# NoteServer

**NoteServer** is a simple RESTful API server implemented in Go for managing notes. It provides endpoints for user authentication, note creation, reading, updating, and deletion, as well as handling multiple notes. Powered by PostgreSQL, it ensures secure data storage and retrieval. The server employs JWT for authentication and integrates the Yandex.Spelling API for accurate spellchecking, enhancing the overall note-taking experience.

## Project description

NoteServer uses the following libraries: 
-   [JWT-GO](github.com/dgrijalva/jwt-go): A library for JSON Web Tokens (JWT) in Go.
-   [Mux](github.com/gorilla/mux): A powerful URL router and dispatcher for building RESTful web services.
-   [PGXv4](github.com/jackc/pgx/v4): A PostgreSQL driver for Go applications.
-   [Logrus](github.com/sirupsen/logrus): A structured logger for Go.
- [Yandex.Speller](https://yandex.ru/dev/speller/#spellcheck): A spellchecking AI-tool API from Yandex.

### Project structure

```
├── Dockerfile
├── LICENSE.md
├── README.md
├── cmd
│   └── noteserver
│       └── main.go
├── go.mod
├── go.sum
└── internal
    └── pkg
        ├── actions
        │   └── action.go
        ├── api
        │   ├── auth.go
        │   ├── handlers.go
        │   └── middlewares.go
        ├── logger
        │   └── setup.go
        ├── models
        │   ├── note.go
        │   ├── spellcheckdata.go
        │   └── user.go
        ├── notes
        │   └── notes.go
        ├── responses
        │   ├── allNotes.go
        │   ├── createUpdateNote.go
        │   ├── deleteNote.go
        │   └── readNote.go
        └── yandex
            └── spellcheck.go
```

## Getting started

### Setting up the infrastructure

**Prerequisites**: 

 - Docker 
 - PostgreSQL client (any)

**Setting up PostgreSQL in the Docker**
Get the PostgreSQL image:
```
docker pull postgres
```
Create the container for PostgreSQL server and run it:
```
docker run --name notes -p 8080:8080 -e POSTGRES_PASSWORD=mysecretpassword -d postgres
```

**Setting up the database**

Connect to the database using client of your choice, for example:  
```
pgcli -h localhost -p 5432 -u postgres
```
Set up the tables:
```
CREATE TABLE Users (
user_id SERIAL PRIMARY KEY,
username VARCHAR(50) NOT NULL,
password_hash VARCHAR(100) NOT NULL,
);

CREATE TABLE Notes (
note_id SERIAL PRIMARY KEY,
user_id INT REFERENCES Users(user_id),
title VARCHAR(100) NOT NULL,
content TEXT,
created_at TIMESTAMP DEFAULT NOW()
);
```

### Setting up the NoteServer

**Cloning a repository**
```
git clone https://github.com/no-to-mediocrity/Noteserver.git
```
**Building a Docker image**  
To build a NoteServer Docker image, run:
```
docker build -t notes:1.0 /Users/user/go/src/noteserver
```
Where `/Users/user/go/src/noteserver` is a path to the project

Then retrieve PostgreSQL server address by running:
  ```
  docker inspect notes | grep IPAddress
```
where `notes` is the name of your PostgreSQL Container

To run  a NoteServer Docker image, use the following command 
```
docker run -d --name note_server -p 8080:8080 \
notes:1.0 \
go/bin/noteserver -sql-server "postgresql://postgres:mysecretpassword@172.17.0.2:5432/postgres"
```
where `mysecretpassword` is the the password of your PostgreSQL Container and `172.17.0.2` is the IP of docker PostgreSQL Container
  
  
## Flags
### --port
**Default**: 8080
**Description**: Specifies the port number on which the server will listen for incoming requests.

**Example usage:**
```
./noteserver --port 8000
```

### --jwt-secret

**Default**: your-secret-key

**Description**: Sets the secret key used for generating and validating JSON Web Tokens (JWTs) for authentication and authorization.

**Example usage:**
```
./noteserver --jwt-secret mysupersecret
```
### --sql-server

**Default**: postgresql://postgres:mysecretpassword@localhost:5432/postgres
**Description**: Defines the connection parameters for the SQL server. Use the appropriate URL format for your SQL server.

**Example usage:**
```
./noteserver --sql-server mysql://user:password@localhost:3306/dbname
```
### --timeout
**Default**: 5

**Description**: Specifies the timeout duration in seconds for Yandex API requests made by the application.

**Example usage:**
```
./noteserver --timeout 10
```

## API endpoints and functionality

Use [Postman Collection](https://api.postman.com/collections/29498342-36cb3529-bd18-4410-87b1-195155e51067?access_key=PMAT-01H9EC868GRK3SBDP58Z3782H3) to test the API . 

**Endpoint**: `/v1/register`

-   **Method**: POST
-   **Purpose**: Allows users to register with a username and password.
-   **Request Body**: JSON containing `"username"` and `"password"` fields.
-  **Response Body**: JSON message.

**Endpoint**: `/v1/login`

-   **Method**: POST
-   **Purpose**: Enables user login using username and password.
-   **Request Body**: JSON containing `"username"` and `"password"` fields.
-   **Response Body**: JSON with an authentication token.

**Endpoint**: `/v1/deleteuser`

-   **Method**: DELETE
-   **Purpose**: Deletes a user account.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token obtained from the login request.
-   **Response Body**: JSON message.


**Endpoint**: `/v1/allnotes`

-   **Method**: GET
-   **Purpose**: Retrieves a list of all notes for the authenticated user.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token.
- **Response Body**: JSON containing `"status"` ,`"message"`, `notes` fields.  


**Endpoint**: `/v1/note`

-   **Method**: POST
-   **Purpose**: Creates a new note with a title and content.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token.
-   **Request Body**: JSON containing `"title"` and `"content"` fields.
-  **Response Body**: JSON containing `"status"` ,`"message"`, `note_id`, `"spelling"`,`"spelling_suggestion"` fields. 


**Endpoint**: `/v1/note`

-   **Method**: GET
-   **Purpose**: Retrieves a specific note by its ID.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token.
-   **Request Body**: JSON containing `"id"` field.
- **Response Body**: JSON containing `"status"` ,`"message"`, `note` fields.  

**Endpoint**: `/v1/note`

-   **Method**: PATCH
-   **Purpose**: Updates an existing note with new title and content.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token.
-  **Response Body**: JSON containing `"status"` ,`"message"`, `note_id`, `"spelling"`,`"spelling_suggestion"` fields. 


**Endpoint**: `/v1/note`

-   **Method**: DELETE
-   **Purpose**: Deletes a specific note by its ID.
-   **Request Headers**: Requires `"Authorization"` header with the authentication token.
-   **Request Body**: JSON containing `"id"` field.
-  **Response Body**: JSON containing `"status"` ,`"message"` fields.

## License 

   **Copyright (C) 
   2023  
   no-to-mediocrity**

  This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
