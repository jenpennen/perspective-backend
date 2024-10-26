package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/joho/godotenv"
)

//go:embed migrations
var migrations embed.FS

func runMigrations() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	migrator, err := migrate.NewMigrator(context.Background(), conn, "public.schema_version")
	if err != nil {
		return err
	}

	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	if err := migrator.LoadMigrations(sub); err != nil {
		return err
	}

	if err := migrator.Migrate(context.Background()); err != nil {
		return err
	}

	return nil
}
func newPool() (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}

type User struct {
	id string
	firstName string
	lastName string
	email string
}
// inserts a new user into the users table and returns the user id
func insertUser(pool *pgxpool.Pool, firstName, lastName, email string) (uuid.UUID, error) {
    var id uuid.UUID

    row := pool.QueryRow(
        context.Background(),
        "insert into users(first_name, last_name, email) values ($1, $2, $3) on conflict(email) do update set email=$3 returning id",
        firstName, lastName, email,
    )

    err := row.Scan(&id)
    if err != nil {
        return uuid.Nil, fmt.Errorf("failed to insert user: %w", err)
    }

    return id, nil
}

// deletes a user in the users table by user id
func deleteUser(pool *pgxpool.Pool, id uuid.UUID) error {
	result, err := pool.Exec(context.Background(), "delete from users where id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no user found with id %v", id)
	}
	return nil

}
// get user by name
func getUser(pool *pgxpool.Pool, column, value string) (*User, error) {
	columns := map[string] bool {
		"first_name" : true,
		"last_name" : true,
		"email" : true,
	}

	if !columns[column] {
		return nil, fmt.Errorf("invalid column name: %s", column)
	}
	query := fmt.Sprintf("select id, first_name, last_name, email from users where %s = $1", column)

	var user User

	row := pool.QueryRow(context.Background(), query, value)
	err := row.Scan(&user.id, &user.firstName, &user.lastName, &user.email)

	if err != nil {
        return nil, fmt.Errorf("failed to retrieve user: %w", err)
    }

    return &user, nil
}

//deletes user by id 
func main() {
	err:= godotenv.Load()
	if err != nil { log.Fatal(err)}
	if err = runMigrations(); err != nil {
		log.Fatal(err)
	}
	pool, err := newPool()
	if err != nil {
		log.Fatal(err)
	}
	// row := pool.QueryRow(context.Background(), "insert into users(first_name, last_name, email) values ($1, $2, $3) returning id", "jenny", "cho", "jennycho35@gmail.com")

	// var id uuid.UUID
	// err = row.Scan(&id)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(id)
 	userID, err := insertUser(pool, "Jenny", "Cho", "jennycho35@gmail.com")
    if err != nil {
        log.Fatalf("Error inserting user: %v\n", err)
    }
	fmt.Printf("New user ID: %v\n", userID)
	err = deleteUser(pool, userID)
	if err != nil {
        log.Printf("Error deleting user: %v\n", err)
    } else {
        fmt.Println("User deleted successfully")
    }
	userID, err = insertUser(pool, "Anish", "Sinha", "anishsinha0128@gmail.com")
    if err != nil {
        log.Fatalf("Error inserting user: %v\n", err)
    }
	fmt.Printf("New user ID: %v\n", userID)	
	userID, err = insertUser(pool, "Jenny", "Kim", "jennykim12@gmail.com")
    if err != nil {
        log.Fatalf("Error inserting user: %v\n", err)
    }
	fmt.Printf("New user ID: %v\n", userID)	
	userID, err = insertUser(pool, "Jenny", "Cho", "jennycho35@gmail.com")
    if err != nil {
        log.Fatalf("Error inserting user: %v\n", err)
    }
	fmt.Printf("New user ID: %v\n", userID)	

	//retrieve user by email
	user, err := getUser(pool,"first_name", "Jenny")
	if err != nil {
        log.Printf("Error retrieving user: %v\n", err)
    } else {
        fmt.Printf("User found: %+v\n", user)
    }

}

