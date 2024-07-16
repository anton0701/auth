package main

import (
	"context"
	"database/sql"
	"github.com/brianvoe/gofakeit"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"

	desc "github.com/anton0701/auth/grpc/pkg/user_v1"
)

const (
	dbDSN = "host=localhost port=54321 dbname=auth user=auth-user password=auth-password"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	builderInsert := sq.Insert("auth").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "role").
		Values(gofakeit.Name(), gofakeit.Email(), 1).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build insert query, error: %s", err)
	}

	var authId int
	err = pool.QueryRow(ctx, query, args...).Scan(&authId)
	if err != nil {
		log.Fatalf("failed to insert using builder, error: %s", err)
	}

	log.Printf("inserted auth with id: %d", authId)

	builderSelect := sq.Select("id", "name", "email", "role", "created_at", "updated_at").
		From("auth").
		PlaceholderFormat(sq.Dollar).
		OrderBy("id").
		Limit(10)

	query, args, err = builderSelect.ToSql()
	if err != nil {
		log.Fatalf("failed to build select query, error: %s", err)
	}

	rows, err := pool.Query(ctx, query, args...)

	log.Printf("Select result:\n\n")
	for rows.Next() {
		var id int
		var name, email string
		var createdAt time.Time
		var updatedAt sql.NullTime
		var role desc.UserRole

		err = rows.Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
		if err != nil {
			log.Fatalf("failed to scan row, error: %s", err)
		}

		log.Printf("id: %v, name: %v, email: %v, createdAt: %v, updatedAt: %v, role: %v", id, name, email, createdAt, updatedAt, role)
	}
}
