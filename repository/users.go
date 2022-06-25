package repository

import (
	"fmt"
	"regexp"

	arango "github.com/arangodb/go-driver"
	"github.com/mandrakey/shoptrac/config"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	COLLECTION_USERS = "users"
	USERLEVEL_ADMIN  = 99
	USERLEVEL_USER   = 0
)

var (
	rxUuid *regexp.Regexp = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
)

type User struct {
	Key      string `json:"_key"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
}

func NewUser() *User {
	return &User{Level: USERLEVEL_USER}
}

func GetUsers(sess *Session) ([]*User, error) {
	if sess.User.Level != USERLEVEL_ADMIN {
		return nil, fmt.Errorf("Must be administrator.")
	}

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

	res := make([]*User, c.Count())
	for {
		u := NewUser()
		_, err := c.ReadDocument(ctx, u)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, u)
	}

	return res, nil
}

func GetUser(uuid string) (*User, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(
		ctx,
		"FOR u IN users FILTER u._key == @uuid RETURN u",
		map[string]interface{}{"uuid": uuid},
	)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	u := NewUser()
	_, err = c.ReadDocument(ctx, u)
	if err != nil {
		return nil, err
	} else {
		return u, nil
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

	u := NewUser()
	_, err = c.ReadDocument(ctx, u)
	if err != nil {
		return nil, err
	} else {
		return u, nil
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

func IsPasswordForUserKey(key string, password string) (bool, error) {
	log := config.Logger()

	db, err := GetDb()
	if err != nil {
		return false, err
	}

	c, err := db.Query(
		ctx,
		"FOR u IN users FILTER u._key == @key RETURN u.password",
		map[string]interface{}{"key": key},
	)
	if err != nil {
		log.Errorf("Failed to query password for user by key: %s", err)
		return false, err
	}
	defer c.Close()

	var dbPassword string
	_, err = c.ReadDocument(ctx, &dbPassword)
	if err != nil {
		if arango.IsNoMoreDocuments(err) {
			return false, nil
		} else {
			log.Errorf("Failed to read password for user '%s': %s", key, err)
			return false, err
		}
	}

	return bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password)) == nil, nil
}

/*
Add the provided [User] to the users collection. The provided [Session] is used to determine, wheter or not the current
user may add new users. On success, this function returns the newly created user's unique key as a string and an error
value of nil. If the current user is not authorized or any other problem occurs, it returns an empty string and an
[error].
*/
func UserAdd(sess *Session, user *User) (string, error) {
	log := config.Logger()

	if sess.User.Level != USERLEVEL_ADMIN {
		log.Warningf("User %s tried to add another user with insufficient privileges.", user.Key)
		return "", fmt.Errorf("Not authorized")
	}

	key, err := uuid.NewV4()
	if err != nil {
		log.Errorf("Failed to generate uuid for new user: %s", err)
		return "", err
	}
	user.Key = key.String()

	err = validateUser(user, false)
	if err != nil {
		log.Errorf("Refused to add new user: %s", err)
		return "", err
	}

	col, err := GetCollection(COLLECTION_USERS)
	if err != nil {
		log.Errorf("Failed to access users collection: %s", err)
		return "", err
	}

	_, err = col.CreateDocument(ctx, user)
	if err != nil {
		log.Errorf("Failed to create new user: %s", err)
		return "", err
	}

	return user.Key, nil
}

/*
UserUpdate is used to update stored information for a specific user. The passed in [Session] is used to determine,
whether or not the current user may use this function. If the current user is not authorized to
change other users, or if any other problem occurs, the function returns an [error]. On success, the return value is
nil.

Note: Only the the email address, the full name, or the user level can be altered using this function. All other
changes to the supplied [User] object are ignored.
*/
func UserUpdate(sess *Session, user *User) error {
	log := config.Logger()

	if sess.User.Level != USERLEVEL_ADMIN && user.Key != sess.User.Key {
		log.Warningf("User %s tried to update another user with insufficient privileges.", user.Key)
		return fmt.Errorf("Not authorized")
	}

	err := validateUser(user, true)
	if err != nil {
		log.Errorf("Refused to update user: %s", err)
		return err
	}

	col, err := GetCollection(COLLECTION_USERS)
	if err != nil {
		log.Errorf("Failed to access users collection: %s", err)
		return err
	}

	data := map[string]interface{}{
		"email": user.Email,
		"name":  user.Name,
		"level": user.Level,
	}
	_, err = col.UpdateDocument(ctx, user.Key, data)
	return err
}

/*
Used to remove a user with the provided unique key from the users collection. The provided [Session] is used to
determine, wether or not the current user may delete other users (which is an administrative function). If the current
user is not authorized to delete users, or in case of other problems, the function returns an [error]. On success, the
return value is nil.
*/
func UserDelete(sess *Session, key string) error {
	log := config.Logger()

	col, err := GetCollection(COLLECTION_USERS)
	if err != nil {
		log.Errorf("Failed to access users collection: %s", err)
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

/*
Updates an existing [User]'s password in the database. The user document is identified using the provided unique key.
On success, the function returns an error value of nil.

Todo: Add configurable password rules and corresponding checks.
*/
func UserUpdatePassword(key string, newpass string) error {
	col, err := GetCollection(COLLECTION_USERS)
	if err != nil {
		return err
	}

	log := config.Logger()
	conf := config.GetAppConfig()

	hashed, err := bcrypt.GenerateFromPassword([]byte(newpass), conf.PasswordCost)
	if err != nil {
		log.Errorf("Failed to hash password entry: %s", err)
		return err
	}

	_, err = col.UpdateDocument(
		ctx,
		key,
		map[string]interface{}{"password": string(hashed)},
	)
	return err
}

func validateUser(u *User, forUpdate bool) error {
	if u.Key == "" {
		return fmt.Errorf("Missing user key")
	}
	if !rxUuid.Match([]byte(u.Key)) {
		return fmt.Errorf("User key must be a UUIDv4 value")
	}
	if u.Email == "" {
		return fmt.Errorf("Email must not be empty")
	}
	if !forUpdate && u.Username == "" {
		return fmt.Errorf("Username must not be empty")
	}
	if !forUpdate && u.Name == "" {
		return fmt.Errorf("User full name must not be empty")
	}
	return nil
}
