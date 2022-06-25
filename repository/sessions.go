package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"
	"time"

	arango "github.com/arangodb/go-driver"
	"github.com/mandrakey/shoptrac/config"
	uuid "github.com/nu7hatch/gouuid"
)

const (
	COLLECTION_SESSIONS      = "sessions"
	DB_TIME_FORMAT           = "2006-01-02 15:04:05Z"
	REMEMBER_ME_TOKEN_CHARS  = "z)2CG5E=ORokg3!4xWaPKMT?;bdZ,DImUiY6+h1.:(qBSXnce7lQ0Ayvpq8st9rNV&FuHjfL-J"
	REMEMBER_ME_TOKEN_LENGTH = 32
)

type Session struct {
	Key               string `json:"_key"`
	UserKey           string `json:"user_key"`
	Created           string `json:"created"`
	Expires           string `json:"expires"`
	RememberMeToken   string `json:"remember_me_token"`
	RememberMeExpires string `json:"remember_me_expires"`
	User              *User  `json:"-"`
}

func NewSession(user *User) (*Session, error) {
	if user == nil {
		return nil, fmt.Errorf("Cannot create session without user data.")
	}

	sess := Session{
		UserKey: user.Key,
		User:    user,
	}

	config := config.GetAppConfig()

	now := time.Now().UTC()
	sess.SetCreatedFromTime(now)
	sess.SetExpiresFromTime(now.Add(time.Minute * time.Duration(config.SessionExpiry)))
	return &sess, nil
}

func GetSessionById(sessionId string) (*Session, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(
		ctx,
		"FOR s IN sessions FILTER s._key == @sessionId AND DATE_DIFF(DATE_NOW(), s.expires, \"s\") >= 0 RETURN s",
		map[string]interface{}{"sessionId": sessionId},
	)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var s Session
	_, err = c.ReadDocument(ctx, &s)
	if err != nil {
		if arango.IsNoMoreDocuments(err) {
			return nil, nil
		}
		return nil, err
	}

	// Add user data
	user, err := GetUser(s.UserKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to load session user: %s", err)
	}
	s.User = user

	return &s, nil
}

func GetSessionWithToken(sessionId string, token string) (*Session, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	log := config.Logger()

	hashed := HashRememberMeToken(token)

	c, err := db.Query(
		ctx,
		"FOR s IN sessions FILTER s._key == @sessionid AND s.remember_me_token == @token AND DATE_DIFF(DATE_NOW(), s.remember_me_expires, \"s\") >= 0 RETURN s",
		map[string]interface{}{"sessionid": sessionId, "token": hashed},
	)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var s Session
	_, err = c.ReadDocument(ctx, &s)
	if arango.IsNoMoreDocuments(err) {
		log.Error("No matching session found.")
		return nil, nil
	} else if err != nil {
		log.Errorf("Failed to read session data: %s", err)
		return nil, err
	}

	// Add user data
	user, err := GetUser(s.UserKey)
	if err != nil {
		return nil, err
	}
	s.User = user

	return &s, nil
}

func AddSession(session *Session) (string, error) {
	col, err := GetCollection(COLLECTION_SESSIONS)
	if err != nil {
		return "", err
	}

	key, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %s", err)
	}
	session.Key = key.String()

	_, err = col.CreateDocument(ctx, session)
	if err != nil {
		return "", nil
	}

	return session.Key, nil
}

func UpdateSession(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_SESSIONS)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeleteSession(key string) error {
	col, err := GetCollection(COLLECTION_SESSIONS)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

func DeleteExpiredSessions() error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	col, err := GetCollection(COLLECTION_SESSIONS)
	if err != nil {
		return err
	}

	c, err := db.Query(
		ctx,
		"FOR s IN sessions FILTER DATE_DIFF(DATE_NOW(), s.expires, \"s\") < 0 return s._key",
		nil,
	)
	if err != nil {
		return err
	}

	for {
		var key string
		_, err = c.ReadDocument(ctx, &key)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}

		_, err = col.RemoveDocument(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateRememberMeToken() (string, string, error) {
	chars := strings.Split(REMEMBER_ME_TOKEN_CHARS, "")
	max := big.NewInt(int64(len(chars) - 1))
	res := make([]string, REMEMBER_ME_TOKEN_LENGTH)
	for i := 0; i < REMEMBER_ME_TOKEN_LENGTH; i++ {
		r, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", "", err
		}
		res[i] = chars[r.Int64()]
	}

	token := strings.Join(res, "")
	sum := HashRememberMeToken(token)
	return token, sum, nil
}

func HashRememberMeToken(token string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
}

func (sess *Session) CreatedAsTime() (*time.Time, error) {
	res, err := time.Parse(DB_TIME_FORMAT, sess.Created)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (sess *Session) SetCreatedFromTime(t time.Time) {
	sess.Created = t.UTC().Format(DB_TIME_FORMAT)
}

func (sess *Session) ExpiresAsTime() (*time.Time, error) {
	res, err := time.Parse(DB_TIME_FORMAT, sess.Expires)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (sess *Session) SetExpiresFromTime(t time.Time) {
	sess.Expires = t.UTC().Format(DB_TIME_FORMAT)
}

func (sess *Session) RememberMeExpiresAsTime() (*time.Time, error) {
	if sess.RememberMeExpires == "" {
		return nil, nil
	}

	res, err := time.Parse(DB_TIME_FORMAT, sess.RememberMeExpires)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (sess *Session) SetRememberMeExpiresFromTime(t time.Time) {
	sess.RememberMeExpires = t.UTC().Format(DB_TIME_FORMAT)
}
