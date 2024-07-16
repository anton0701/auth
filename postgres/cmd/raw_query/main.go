package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v4"

	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

const (
	dbDSN = "host=localhost port=54321 dbname=auth user=auth-user password=auth-password"
)

func main() {
	ctx := context.Background()

	con, err := pgx.Connect(ctx, dbDSN)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer con.Close(ctx)

	res, err := con.Exec(ctx, "INSERT INTO auth (name, email, role) VALUES ($1, $2, $3)", gofakeit.Name(), gofakeit.Email(), 1)
	if err != nil {
		log.Fatalf("failed to insert in auth: %v", err)
	}

	log.Printf("inserted %d rows", res.RowsAffected())

	rows, err := con.Query(ctx, "SELECT id, name, email, role, created_at, updated_at FROM auth")
	if err != nil {
		log.Fatalf("failed to select from auth: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		var createdAt time.Time
		var updatedAt sql.NullTime
		var role desc.UserRole

		err := rows.Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
		if err != nil {
			log.Fatalf("failed to scan auth: %v", err)
		}

		log.Printf("id: %v, name: %v, email: %v, createdAt: %v, updatedAt: %v, role: %v", id, name, email, createdAt, updatedAt, role)
	}
}
