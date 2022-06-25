package handler

import (
	"encoding/json"
	"time"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

const (
	CONTEXT_KEY_SESSION = "session"
)

func UnauthorizedResponse() (int, string) {
	return 401, ""
}

func ErrorResponse(message string) string {
	return ErrorResponseWithData(message, nil)
}

func ErrorResponseWithData(message string, data interface{}) string {
	obj := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	res, err := json.Marshal(obj)
	if err != nil {
		config.Logger().Errorf("Failes to generate error response: %s", err)
		return ""
	}

	return string(res)
}

func SuccessResponse(data interface{}) string {
	res, err := json.Marshal(data)
	if err != nil {
		config.Logger().Errorf("Failed to generate success response: %s", err)
		return ""
	}

	return string(res)
}

func ExtractSessionIdFromHeader(ctx *macaron.Context) string {
	authValue := ctx.Req.Header["Authentication"]
	if len(authValue) == 0 {
		return ""
	} else {
		return authValue[0]
	}
}

func GetActiveSession(ctx *macaron.Context) *repository.Session {
	if ctx.Data[CONTEXT_KEY_SESSION] == nil {
		return nil
	} else {
		log := config.Logger()
		sess, ok := ctx.Data[CONTEXT_KEY_SESSION].(*repository.Session)
		if !ok {
			log.Error("Failed to obtain Session from context.")
			return nil
		} else {
			return sess
		}
	}
}

func IsValidSession(ctx *macaron.Context) bool {
	log := config.Logger()
	if ctx.Data[CONTEXT_KEY_SESSION] == nil {
		return false
	}

	sess, ok := ctx.Data[CONTEXT_KEY_SESSION].(*repository.Session)
	if !ok {
		log.Error("Failed to obtain Session from context.")
		return false
	}
	now := time.Now().UTC()

	created, err := sess.CreatedAsTime()
	if err != nil || created == nil {
		log.Errorf("Failed to parse session created time: %s", err)
		return false
	}
	if created.After(now) {
		log.Debugf("Created (%s) is AFTER now (%s).", sess.Created, now.Format(repository.DB_TIME_FORMAT))
		return false
	}

	expires, err := sess.ExpiresAsTime()
	if err != nil || expires == nil {
		log.Errorf("Failed to parse session expires time: %s", err)
		return false
	}
	if expires.Before(now) {
		log.Debugf("Expires (%s) is BEFORE now (%s).", sess.Expires, now.Format(repository.DB_TIME_FORMAT))
		return false
	}

	return true
}
