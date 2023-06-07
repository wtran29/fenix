//go:build integration

// run tests with this command: go test -v ./data --tags integration --count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "secret"
	dbName   = "fenix_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Test",
	LastName:  "Dummy",
	Email:     "me@here.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")
	os.Setenv("UPPER_DB_LOG", "ERROR")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	pool = p
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("Error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker : %s", err)
	}

	err = createTables(testDB)
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	models = New(testDB)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	stmt := `
	CREATE OR REPLACE FUNCTION trigger_set_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	drop table if exists users cascade;

	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		first_name character varying(255) NOT NULL,
		last_name character varying(255) NOT NULL,
		user_active integer NOT NULL DEFAULT 0,
		email character varying(255) NOT NULL UNIQUE,
		password character varying(60) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);

	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();

	drop table if exists remember_tokens;

	CREATE TABLE remember_tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		remember_token character varying(100) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);

	CREATE TRIGGER set_timestamp
	BEFORE UPDATE ON remember_tokens
	FOR EACH ROW
	EXECUTE PROCEDURE trigger_set_timestamp();

	drop table if exists tokens;

	CREATE TABLE tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		first_name character varying(255) NOT NULL,
		email character varying(255) NOT NULL,
		token character varying(255) NOT NULL,
		token_hash bytea NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now(),
		expiry timestamp without time zone NOT NULL
	);

	CREATE TRIGGER set_timestamp
	BEFORE UPDATE ON tokens
	FOR EACH ROW
	EXECUTE PROCEDURE trigger_set_timestamp();
		`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table name returned: ", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert user: ", err)
	}

	if id == 0 {
		t.Error("0 returned as id after insertion")
	}
}

func TestUser_Get(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get all users: ", err)
	}

}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0: ", err)
	}
}

func TestUser_Update(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	u.LastName = "Monster"
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to update user: ", err)
	}

	u, err = models.Users.Get(1)
	if err != nil {
		t.Error("failed to get updated user: ", err)
	}

	if u.LastName != "Monster" {
		t.Error("last name not updated in db")
	}
}

func TestUser_IsPasswordMatch(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}
	matches, err := u.IsPasswordMatch("password")
	if err != nil {
		t.Error("error checking password: ", err)
	}

	if !matches {
		t.Error("password does not match but should")
	}

	matches, err = u.IsPasswordMatch("123")
	if err != nil {
		t.Error("error checking password: ", err)
	}

	if matches {
		t.Error("password match but should not")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	err := models.Users.ResetPassword(1, "new_password")
	if err != nil {
		t.Error("error resetting password: ", err)
	}

	err = models.Users.ResetPassword(2, "new_password")
	if err == nil {
		t.Error("did not get an error when trying to reset password for non-existent user")
	}
}

func TestUser_Delete(t *testing.T) {
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete user: ", err)
	}

	_, err = models.Users.Get(1)
	if err == nil {
		t.Error("retrieved user but supposed to be deleted: ", err)
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Error("wrong table name returned for tokens")
	}
}

func TestToken_GenerateToken(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("error inserting user: ", err)
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token: ", err)
	}
}

func TestToken_Insert(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token: ", err)
	}

	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("error inserting token: ", err)
	}
}

func TestToken_GetUserByToken(t *testing.T) {
	token := "abc"
	_, err := models.Tokens.GetUserByToken(token)
	if err == nil {
		t.Error("error expected but not received when getting user with bad token: ", err)
	}

	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	_, err = models.Tokens.GetUserByToken(u.Token.PlainText)
	if err != nil {
		t.Error("failed to get user with valid token: ", err)
	}
}

func TestToken_GetUserTokens(t *testing.T) {
	tokens, err := models.Tokens.GetUserTokens(1)
	if err != nil {
		t.Error(err)
	}

	if len(tokens) > 0 {
		t.Error("tokens returned for non-existent user")
	}
}

func TestToken_GetTokenByID(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user: ", err)
	}
	_, err = models.Tokens.GetTokenByID(u.Token.ID)
	if err != nil {
		t.Error("error getting token by id: ", err)
	}
}

func TestToken_GetToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user: ", err)
	}
	_, err = models.Tokens.GetToken(u.Token.PlainText)
	if err != nil {
		t.Error("error getting token by plaintext token: ", err)
	}

	_, err = models.Tokens.GetToken("123")
	if err == nil {
		t.Error("error getting non-existent token by plaintext token: ", err)
	}
}

var authData = []struct {
	name          string
	token         string
	email         string
	errorExpected bool
	message       string
}{
	{"invalid", "abcdefghijklmnopqrstuv", "a@here.com", true, "invalid - token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxy", "a@here.com", true, "wrong token length - token accepted as valid"},
	{"no_user", "abcdefghijklmnopqrstuv", "a@here.com", true, "no user - token accepted as valid"},
	{"valid", "", "me@here.com", false, "valid token reported as invalid"},
}

func TestToken_AuthenticateToken(t *testing.T) {
	for _, tt := range authData {
		token := ""
		if tt.email == dummyUser.Email {
			user, err := models.Users.GetByEmail(tt.email)
			if err != nil {
				t.Error("failed to get user: ", err)
			}
			token = user.Token.PlainText
		} else {
			token = tt.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+token)

		_, err := models.Tokens.AuthenticateToken(req)

		if tt.errorExpected && err == nil {
			t.Errorf("%s: %s", tt.name, tt.message)
		} else if !tt.errorExpected && err != nil {
			t.Errorf("%s: %s - %s", tt.name, tt.message, err)
		} else {
			t.Logf("passed %s", tt.name)
		}
	}
}

func TestToken_Delete(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error(err)
	}

	err = models.Tokens.DeleteToken(u.Token.PlainText)
	if err != nil {
		t.Error("error deleting token: ", err)
	}
}

func TestToken_ExpiredToken(t *testing.T) {
	// insert a token
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email: ", err)
	}

	token, err := models.Tokens.GenerateToken(u.ID, -time.Hour)
	if err != nil {
		t.Error("failed to generate token: ", err)
	}

	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("failed to insert token: ", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)

	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("failed to catch expired token")
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	_, err := models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("failed to catch missing auth header")
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "abc")
	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("failed to catch bad auth header")
	}

	newUser := User{
		FirstName: "temp",
		LastName:  "user",
		Email:     "you@there.com",
		Active:    1,
		Password:  "abc",
	}

	id, err := models.Users.Insert(newUser)
	if err != nil {
		t.Error("failed to insert a new user: ", err)
	}

	token, err := models.Tokens.GenerateToken(id, 1*time.Hour)
	if err != nil {
		t.Error("failed to generate token: ", err)
	}

	err = models.Tokens.Insert(*token, newUser)
	if err != nil {
		t.Error("failed to insert token: ", err)
	}

	err = models.Users.Delete(id)
	if err != nil {
		t.Error("failed to delete user: ", err)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)
	_, err = models.Tokens.AuthenticateToken(req)
	if err == nil {
		t.Error("failed to catch token for deleted user")
	}

}

func TestToken_DeleteNonExistToken(t *testing.T) {
	err := models.Tokens.DeleteToken("abc")
	if err != nil {
		t.Error("error deleting token")
	}
}

func TestToken_ValidToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email: ", err)
	}

	newToken, err := models.Tokens.GenerateToken(u.ID, 24*time.Hour)
	if err != nil {
		t.Error("failed to generate token: ", err)
	}

	err = models.Tokens.Insert(*newToken, *u)
	if err != nil {
		t.Error("failed to insert token: ", err)
	}

	ok, err := models.Tokens.ValidToken(newToken.PlainText)
	if err != nil {
		t.Error("error calling ValidToken: ", err)
	}

	if !ok {
		t.Error("valid token reported as invalid")
	}

	ok, err = models.Tokens.ValidToken("abc")
	if ok {
		t.Error("invalid token reported as valid")
	}

	u, err = models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email: ", err)
	}

	err = models.Tokens.Delete(u.Token.ID)
	if err != nil {
		t.Error("failed to delete token: ", err)
	}

	ok, err = models.Tokens.ValidToken(u.Token.PlainText)
	if err == nil {
		t.Error(err)
	}

	if ok {
		t.Error("no error reported when validating non-existent token")
	}

}
