/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package handler

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

func UsersGet(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()

	sess := GetActiveSession(ctx)
	if sess.User.Level != repository.USERLEVEL_ADMIN {
		log.Warningf("Illegal access to UsersGet by user %s.", sess.UserKey)
		return 403, ""
	}

	users, err := repository.GetUsers(sess)
	if err != nil {
		log.Errorf("Failed to retrieve user list: %s", err)
		return 500, ErrorResponse("Failed to retrieve user list")
	}

	return 200, SuccessResponse(users)
}

func UsersGetUser(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)
	if sess.User.Level != repository.USERLEVEL_ADMIN {
		log.Warningf("Illegal access to UsersGetUser by user %s", sess.UserKey)
		return 403, ""
	}

	uuid := ctx.Params(":uuid")
	if uuid == "" {
		return 400, ErrorResponse("User UUID must be provided.")
	}

	u, err := repository.GetUser(uuid)
	if err != nil {
		log.Errorf("Failed to retrieve user for key %s: %s", uuid, err)
		return 500, ErrorResponse("Failed to find a user for the provided key")
	}

	return 200, SuccessResponse(u)
}

func UsersPut(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)
	if sess.User.Level != repository.USERLEVEL_ADMIN {
		log.Warningf("Illegal access to UsersPut by user %s", sess.UserKey)
		return 403, ""
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("Failed to read request: %s", err)
		return 500, ErrorResponse("Failed to read request.")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to parse request JSON: %s", err)
		return 500, ErrorResponse("Failed to parse request JSON.")
	}

	// Extract data
	user := repository.NewUser()

	username, ok := data["username"].(string)
	if !ok || username == "" {
		return 400, ErrorResponse("Parameter 'username' is required and must be a non-empty string.")
	}
	user.Username = username

	name, ok := data["name"].(string)
	if !ok || name == "" {
		return 400, ErrorResponse("Parameter 'name' is required and must be a non-empty string.")
	}
	user.Name = name

	email, ok := data["email"].(string)
	if !ok || email == "" {
		return 400, ErrorResponse("Parameter 'email' is required and must be a non-empty string.")
	}
	user.Email = email

	level, ok := data["level"].(int)
	if !ok {
		level = repository.USERLEVEL_USER
	}
	user.Level = level

	password64, ok := data["password"].(string)
	if !ok {
		return 400, ErrorResponse("Parameter 'password' is required and must be a base64 encoded, non-empty string.")
	}
	passwordBytes, err := base64.StdEncoding.DecodeString(password64)
	if err != nil {
		log.Errorf("Failed to decode base64 value for new user password: %s", err)
		return 400, ErrorResponse("Invalid value for 'password'.")
	}

	confirmation64, ok := data["confirmation"].(string)
	if !ok {
		return 400, ErrorResponse("Parameter 'confirmation' is required and must be a base64 encoded, non-empty string.")
	}
	confirmationBytes, err := base64.StdEncoding.DecodeString(confirmation64)
	if err != nil {
		log.Errorf("Failed to decode base64 value for new user password confirmation: %s", err)
		return 400, ErrorResponse("Invalid value for 'confirmation'.")
	}

	if string(passwordBytes) != string(confirmationBytes) {
		return 400, ErrorResponse("Password does not match confirmation.")
	}

	// Store user
	key, err := repository.UserAdd(sess, user)
	if err != nil {
		return 500, ErrorResponse("Failed to add new user.")
	}
	user.Key = key

	err = repository.UserUpdatePassword(key, string(passwordBytes))
	if err != nil {
		log.Warningf("Failed to set initially provided user password for user %s: %s", key, err)
		return 500, ErrorResponse("The user has been created, but the provided password could not be set. Logging in may not be possible without a further password change.")
	}

	return 200, SuccessResponse(user)
}

func UsersPatch(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)
	if sess.User.Level != repository.USERLEVEL_ADMIN {
		log.Warningf("Illegal access to UsersPatch by user %s", sess.UserKey)
		return 403, ""
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("Failed to read request: %s", err)
		return 500, ErrorResponse("Failed to read request.")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to parse request JSON: %s", err)
		return 500, ErrorResponse("Failed to parse request JSON.")
	}

	// Extract data
	user := repository.NewUser()

	key, ok := data["_key"].(string)
	if !ok || key == "" {
		return 400, ErrorResponse("Parameter '_key' is required and must be a non-empty string.")
	}
	user.Key = key

	name, ok := data["name"].(string)
	if !ok || name == "" {
		return 400, ErrorResponse("Parameter 'name' is required and must be a non-empty string.")
	}
	user.Name = name

	email, ok := data["email"].(string)
	if !ok || email == "" {
		return 400, ErrorResponse("Parameter 'email' is required and must be a non-empty string.")
	}
	user.Email = email

	user.Level = repository.USERLEVEL_USER
	levelStr, ok := data["level"].(string)
	if ok {
		level64, err := strconv.ParseInt(levelStr, 10, 0)
		if err == nil {
			user.Level = int(level64)
		}
	}

	var passwordBytes []byte
	var confirmationBytes []byte

	password64, ok := data["password"].(string)
	if ok {
		// Got a new password to set
		passwordBytes, err = base64.StdEncoding.DecodeString(password64)
		if err != nil {
			log.Errorf("Failed to decode base64 value for new user password: %s", err)
			return 400, ErrorResponse("Invalid value for 'password'.")
		}

		confirmation64, ok := data["confirmation"].(string)
		if !ok {
			return 400, ErrorResponse("Parameter 'confirmation' is required and must be a base64 encoded, non-empty string.")
		}
		confirmationBytes, err = base64.StdEncoding.DecodeString(confirmation64)
		if err != nil {
			log.Errorf("Failed to decode base64 value for new user password confirmation: %s", err)
			return 400, ErrorResponse("Invalid value for 'confirmation'.")
		}

		if string(passwordBytes) != string(confirmationBytes) {
			return 400, ErrorResponse("Password does not match confirmation.")
		}
	}

	// Update user
	err = repository.UserUpdate(sess, user)
	if err != nil {
		return 500, ErrorResponse("Failed to update user data.")
	}

	if passwordBytes != nil {
		err = repository.UserUpdatePassword(key, string(passwordBytes))
		if err != nil {
			log.Warningf("Failed to update user password for user %s: %s", key, err)
			return 500, ErrorResponse("The user has been updated, but the provided password could not be set.")
		}
	}

	return 204, ""
}

func UsersDelete(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)
	if sess.User.Level != repository.USERLEVEL_ADMIN {
		log.Warningf("Illegal access to UsersDelete by user %s", sess.UserKey)
		return 403, ""
	}

	uuid := ctx.Params(":uuid")
	if uuid == "" {
		return 400, ErrorResponse("User UUID must be provided.")
	}

	err := repository.UserDelete(sess, uuid)
	if err != nil {
		return 500, ErrorResponse("Failed to remove the specified user.")
	}
	return 204, ""
}

func UsersOptions(ctx *macaron.Context) (int, string) {
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Methods",
		"GET, PATCH, PUT, DELETE, OPTIONS",
	)
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Headers",
		"Content-Type, Authentication",
	)
	return 204, ""
}
