#!/bin/bash

#Docker postgress settings
POSTGRES_PASSWORD="mysecretpassword"
POSTGRES_CONTAINER="notes_database"
POSTGRES_PORT="5432"
NOTESERVER_PORT="8081"
NOTESERVER_CONTAINER="notes_server"

 
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker before running this script."
    exit 1
fi
 
# Get the PostgreSQL image
docker pull postgres
 
# Create and run PostgreSQL container
docker run --name $POSTGRES_CONTAINER -p $POSTGRES_PORT:$POSTGRES_PORT -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD -d postgres
 

sleep 10
 
# Setting up the database
# Connect to the PostgreSQL database using pgcli or any other client
# Replace the hostname and port with appropriate values
 
# Set up tables using pgcli or other PostgreSQL client


SETDB_COMMAND=$(cat <<EOF
CREATE TABLE Users (
  user_id SERIAL PRIMARY KEY,
  username VARCHAR(50) NOT NULL,
  password_hash VARCHAR(100) NOT NULL
);
 
CREATE TABLE Notes (
  note_id SERIAL PRIMARY KEY,
  user_id INT REFERENCES Users(user_id),
  title VARCHAR(100) NOT NULL,
  content TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
EOF
)
 
PGPASSWORD="$POSTGRES_PASSWORD" docker exec -it "$POSTGRES_CONTAINER" psql -U postgres -d postgres -c "$SETDB_COMMAND"
 
# Build NoteServer Docker image
docker build -t notes:1.0 $PWD
 
# Get PostgreSQL container IP address
POSTGRES_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $POSTGRES_CONTAINER)
 
if [ -z "$POSTGRES_IP" ]; then
  echo "Error: can not obtain postgres database IP"
  exit 1
fi

# Run NoteServer Docker image
docker run -d --name $NOTESERVER_CONTAINER -p $NOTESERVER_PORT:8080 \
  notes:1.0 \
  go/bin/noteserver -sql-server "postgresql://postgres:$POSTGRES_PASSWORD@$POSTGRES_IP:$POSTGRES_PORT/postgres"

 