package handler

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

func GetLogout(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()

	sess := GetActiveSession(ctx)
	if sess == nil {
		return 500, ErrorResponse("")
	}

	err := repository.DeleteSession(sess.Key)
	if err != nil {
		log.Errorf("Failed to end/delete session %s: %s", sess.Key, err)
		return 500, ErrorResponse("Failed to end the session.")
	} else {
		return 204, ""
	}
}

func PostLogin(ctx *macaron.Context) (int, string) {
	if IsValidSession(ctx) {
		return 400, ErrorResponse("Already logged in")
	}

	log := config.Logger()
	config := config.GetAppConfig()

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("Failed to read request body: %s", err)
		return 500, ErrorResponse("Failed to read request.")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to parse request JSON: %s", err)
		return 500, ErrorResponse("Failed to parse request JSON.")
	}

	username, ok := data["username"].(string)
	if !ok {
		log.Error("Failed to extract username from request data as string.")
		return 400, ErrorResponse("Invalid value for 'username'.")
	}

	passwordB64, ok := data["password"].(string)
	if !ok {
		log.Error("Failed to extract password from request data as string.")
		return 400, ErrorResponse("Invalid value for 'password'.")
	}

	passwordBytes, err := base64.StdEncoding.DecodeString(passwordB64)
	if err != nil {
		log.Errorf("Failed to decode base64 password: %s", err)
		return 500, ErrorResponse("Invalid value for 'password'.")
	}
	password := string(passwordBytes)

	// ----
	// Check password

	correct, err := repository.IsPasswordForUser(username, password)
	if err != nil {
		log.Errorf("Failed to check provided username/password combination: %s", err)
		return 500, ErrorResponse("Failed to check provided username/password combination.")
	}
	if !correct {
		log.Warningf("Invalid login attempt for user '%s'.", username)
		return UnauthorizedResponse()
	}

	// ----
	// Create and store session

	// Load user data
	user, err := repository.GetUserByUsername(username)
	if err != nil {
		log.Errorf("Failed to retrieve user dataset for '%s': %s", username, err)
		return 500, ErrorResponse("Failed to retrieve user data.")
	}

	// Create new session
	sess, err := repository.NewSession(user)
	if err != nil {
		log.Errorf("Failed to generate new session for '%s': %s", username, err)
		return 500, ErrorResponse("Failed to generate session.")
	}

	rememberMe, ok := data["remember_me"].(bool)
	if !ok {
		rememberMe = false
	}

	baseTime := time.Now()
	rememberMeToken := ""
	rememberMeRawToken := ""
	rememberMeExpiresTime := baseTime
	rememberMeExpiresTimeText := ""
	if rememberMe {
		rememberMeRawToken, rememberMeToken, err = repository.GenerateRememberMeToken()
		if err != nil {
			log.Warningf("Failed to generate remember me token.")
		}
		created, err := sess.CreatedAsTime()
		if err != nil {
			log.Warningf("Failed to set remember expiry, could not extract created time from session: %s", err)
		}

		rememberMeExpiresTime = created.Add(time.Minute * time.Duration(config.SessionRememberMeExpiry))
	}

	if rememberMeToken != "" && rememberMeRawToken != "" && rememberMeExpiresTime != baseTime {
		sess.RememberMeToken = rememberMeToken
		sess.SetRememberMeExpiresFromTime(rememberMeExpiresTime)
		rememberMeExpiresTimeText = rememberMeExpiresTime.Format(repository.DB_TIME_FORMAT)
	}

	// Store session
	uuid, err := repository.AddSession(sess)
	if err != nil {
		log.Errorf("Failed to create session for '%s': %s", username, err)
		return 500, ErrorResponse("Failed to create session.")
	}

	res := map[string]interface{}{
		"uuid":                uuid,
		"expires":             sess.Expires,
		"remember_me_token":   rememberMeRawToken,
		"remember_me_expires": rememberMeExpiresTimeText,
	}
	resString, err := json.Marshal(res)
	if err != nil {
		repository.DeleteSession(uuid)
		log.Errorf("Failed to format session creation response: %s", err)
		return 500, ErrorResponse("Failed to create session.")
	}

	return 200, string(resString)
}

func PostContinue(ctx *macaron.Context) (int, string) {
	if IsValidSession(ctx) {
		return 400, ErrorResponse("Already logged in")
	}

	log := config.Logger()
	config := config.GetAppConfig()

	sessionId := ExtractSessionIdFromHeader(ctx)
	if sessionId == "" {
		log.Warningf("Tried to continue session without session id")
		return 400, ErrorResponse("Cannot continue session without session id.")
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("Failed to read request body: %s", err)
		return 500, ErrorResponse("Failed to read request.")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed parse request JSON: %s", err)
		return 500, ErrorResponse("Failed to parse request JSON.")
	}

	token, ok := data["token"].(string)
	if !ok {
		log.Error("Failed to extract remember me token from request data as string.")
		return 400, ErrorResponse("Invalid value for 'token'.")
	}

	// ----
	// Check token

	sess, err := repository.GetSessionWithToken(sessionId, token)
	if sess == nil || err != nil {
		log.Errorf("Failed to retrieve session '%s' with token: '%s'; Error: %s", sessionId, token, err)
		return 500, ErrorResponse("Failed to retrieve session.")
	}

	// Reset expiry
	sess.SetExpiresFromTime(time.Now().UTC().Add(time.Minute * time.Duration(config.SessionExpiry)))
	sess.SetRememberMeExpiresFromTime(time.Now().UTC().Add(time.Minute * time.Duration(config.SessionRememberMeExpiry)))

	err = repository.UpdateSession(
		sess.Key,
		&map[string]interface{}{
			"expires":             sess.Expires,
			"remember_me_expires": sess.RememberMeExpires,
		},
	)
	if err != nil {
		log.Errorf("Failed to store updated session timeouts for session %s: %s", sessionId, err)
		return 500, ErrorResponse("Failed to store updated session timeouts.")
	}

	return 204, ""
}

func OptionsAuth(ctx *macaron.Context) (int, string) {
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Methods",
		"GET, POST, OPTIONS",
	)
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Headers",
		"Content-Type, Authentication",
	)
	return 200, ""
}
