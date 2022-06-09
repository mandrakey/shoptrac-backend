package repository

import (
	arango "github.com/arangodb/go-driver"
	"github.com/mandrakey/shoptrac/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	COLLECTION_USERS = "users"
)

type User struct {
	Key      string `json:"_key"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func GetUsers() (*[]User, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	// ----
	// Query database

	c, err := db.Query(ctx, "FOR u IN users RETURN u", nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res := make([]User, c.Count())
	for {
		var u User
		_, err := c.ReadDocument(ctx, &u)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, u)
	}

	return &res, nil
}

func GetUser(uuid string) (*User, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(
		ctx,
		"FOR u IN users FILTER u.uuid == @uuid RETURN u",
		map[string]interface{}{"uuid": uuid},
	)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var u User
	_, err = c.ReadDocument(ctx, &u)
	if err != nil {
		return nil, err
	} else {
		return &u, nil
	}
}

func GetUserByUsername(username string) (*User, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(
		ctx,
		"FOR u IN users FILTER u.username == @username RETURN u",
		map[string]interface{}{"username": username},
	)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var u User
	_, err = c.ReadDocument(ctx, &u)
	if err != nil {
		return nil, err
	} else {
		return &u, nil
	}
}

func IsPasswordForUser(username string, password string) (bool, error) {
	log := config.Logger()

	db, err := GetDb()
	if err != nil {
		return false, err
	}

	c, err := db.Query(
		ctx,
		"FOR u IN users FILTER u.username == @username RETURN u.password",
		map[string]interface{}{"username": username},
	)
	if err != nil {
		log.Errorf("Failed to query user data: %s", err)
		return false, err
	}
	defer c.Close()

	var dbPassword string
	_, err = c.ReadDocument(ctx, &dbPassword)
	if err != nil {
		if arango.IsNoMoreDocuments(err) {
			return false, nil
		} else {
			log.Errorf("Failed to read password entry: %s", err)
			return false, err
		}
	}

	return bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)) == nil, nil
}
