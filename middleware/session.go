package middleware

import (
	"time"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/handler"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

func SessionMiddleware() macaron.Handler {
	return func(ctx *macaron.Context) {
		log := config.Logger()
		config := config.GetAppConfig()

		sessionId := handler.ExtractSessionIdFromHeader(ctx)
		sess, err := repository.GetSessionById(sessionId)
		if err != nil {
			log.Errorf("Failed to load session: %s\r\n", err)
			return
		}
		if sess == nil {
			return
		}

		ctx.Data[handler.CONTEXT_KEY_SESSION] = sess

		// Prolong the session
		updateData := make(map[string]interface{})

		now := time.Now().UTC()
		newExpires := now.Add(time.Minute * time.Duration(config.SessionExpiry))
		updateData["expires"] = newExpires.Format(repository.DB_TIME_FORMAT)

		if sess.RememberMeExpires != "" {
			newRememberMeExpires := now.Add(time.Minute * time.Duration(config.SessionRememberMeExpiry))
			updateData["remember_me_expires"] = newRememberMeExpires.Format(repository.DB_TIME_FORMAT)
		}

		err = repository.UpdateSession(sess.Key, &updateData)
		if err != nil {
			log.Warningf("Failed to prolong session: %s", err)
		}
	}
}
