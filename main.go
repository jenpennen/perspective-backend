package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
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
	ID string `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
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
// delete for specific columns ex. DeleteByEmail, DeleteByID
func deleteUserByID(pool *pgxpool.Pool, userID uuid.UUID) error {
	result, err := pool.Exec(context.Background(), "delete from users where id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no user found with id %v", userID)
	}
	return nil

}
// get user by name 
// only return one user here
// create selects for specific columns ex. getUsersByFirstName, getUsersByEmail etc
func getUsersByEmail(pool *pgxpool.Pool, column string, value string) ([]User, error) {
	columns := map[string] bool {
		"first_name" : false,
		"last_name" : false,
		"email" : true,
	}

	if !columns[column] {
		return nil, fmt.Errorf("invalid column name: %s", column)
	}
	// no interpolating in any SQL!!
	query := fmt.Sprintf("select id, first_name, last_name, email from users where %s = $1", column)
	fmt.Println(query)

	rows, err := pool.Query(context.Background(), query, value)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		// row := pool.QueryRow(context.Background(), query, value)
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)

		if err != nil {
        return nil, fmt.Errorf("failed to retrieve user: %w", err)
    	}
		users = append(users, user)
	}
	
	if rows.Err() != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", rows.Err())
	}

    return users, nil
}

//insert test users
func insertTestUser(pool *pgxpool.Pool, firstName, lastName, email string) {
    userID, err := insertUser(pool, firstName, lastName, email)
    if err != nil {
        log.Printf("Error inserting user %s %s: %v\n", firstName, lastName, err)
        return
    }
    fmt.Printf("New user ID for %s %s: %v\n", firstName, lastName, userID)
}

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
	app := newAPIServer(pool).build()
	srv := &http.Server{Addr: ":8890", Handler: app}
	log.Fatal(srv.ListenAndServe())

	testUsers := []struct {
		firstName string
		lastName string
		email string
	}{
		{"Anish", "Sinha", "anishsinha0128@gmail.com"},
        {"Jenny", "Kim", "jennykim12@gmail.com"},
	    {"Jenny", "Cho", "jennycho35@gmail.com"},
		{"Toffee", "Sinha", "toffee123@gmail.com"},
		{"Meadow", "Sinha", "meadow12@gmail.com"},
		{"Melody", "Cho", "melodyc12@gmail.com"},
	    {"Earl", "Cho", "earlthegrey12@gmail.com"},
		{"Honey", "Cho", "honeyb12@gmail.com"},
		{"Almond", "Cho", "almond#1@gmail.com"},
	}
	for _, u := range testUsers {
        insertTestUser(pool, u.firstName, u.lastName, u.email)
	}
	userID, err := insertUser(pool, "firstName", "lastName", "email")
	if err != nil {
        log.Fatalf("Error inserting test user: %v\n", err)
    }
    fmt.Printf("Inserted test user with ID: %v\n", userID)
	fmt.Println("Deleting the test user...")
    err = deleteUserByID(pool, userID)
    if err != nil {
        log.Fatalf("Error deleting test user: %v\n", err)
    }


	// //retrieve user by last name
	// lastName := "Cho"
	// fmt.Printf("Retrieving users based on %v\n", lastName)
	// user, err := getUser(pool,"last_name", lastName)
	// if err != nil {
    //     log.Printf("Error retrieving user: %v\n", err)
    // } else {
    //     fmt.Printf("User found: %v\n", user)
    // }

}

