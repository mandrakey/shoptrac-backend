/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package handler

import (
	"encoding/base64"
	"encoding/json"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

func GetProfile(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)

	user, err := repository.GetUser(sess.UserKey)
	if err != nil {
		log.Errorf("Failed to load user data for GetProfile: %s", err)
		return 500, ErrorResponse("Failed to load user data.")
	}

	return 200, SuccessResponse(user)
}

func PostProfileUpdatePassword(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)

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

	// old password
	oldPassword64, ok := data["old_password"].(string)
	if !ok {
		log.Error("Failed to extract old password from data as string.")
		return 400, ErrorResponse("Invalid value for 'old_password'.")
	}
	oldPasswordBytes, err := base64.StdEncoding.DecodeString(oldPassword64)
	if err != nil {
		log.Errorf("Failed to decode base64 old password: %s", err)
		return 400, ErrorResponse("Invalid value for 'old_password'.")
	}
	oldPassword := string(oldPasswordBytes)

	correct, err := repository.IsPasswordForUserKey(sess.UserKey, oldPassword)
	if err != nil {
		log.Errorf("Failed to check provided old password: %s", err)
		return 500, ErrorResponse("Failed to check old password.")
	}
	if !correct {
		return 400, ErrorResponse("The provided password was not valid.")
	}

	// new password
	newPassword64, ok := data["password"].(string)
	if !ok {
		log.Error("Failed to extract new password from data as string.")
		return 400, ErrorResponse("Invalid value for 'password'.")
	}
	newPasswordBytes, err := base64.StdEncoding.DecodeString(newPassword64)
	if err != nil {
		log.Errorf("Failed to decode base64 new password: %s", err)
		return 400, ErrorResponse("Invalid value for 'password'.")
	}
	newPassword := string(newPasswordBytes)

	// confirmation
	confirmation64, ok := data["confirmation"].(string)
	if !ok {
		log.Error("Failed to extract confirmation from data as string.")
		return 400, ErrorResponse("Invalid value for 'confirmation'.")
	}
	confirmationBytes, err := base64.StdEncoding.DecodeString(confirmation64)
	if err != nil {
		log.Errorf("Failed to decode base64 password confirmation: %s", err)
		return 400, ErrorResponse("Invalid value for 'confirmation'.")
	}
	confirmation := string(confirmationBytes)

	if newPassword != confirmation {
		return 400, ErrorResponse("The new password and it's confirmation do not match.")
	}

	// ----
	// Update

	err = repository.UserUpdatePassword(sess.UserKey, newPassword)
	if err == nil {
		return 204, ""
	} else {
		return 500, ErrorResponse("Failed to update user password.")
	}
}

func PatchProfile(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()
	sess := GetActiveSession(ctx)

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("Failed to request body: %s", err)
		return 500, ErrorResponse("Failed to read request.")
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Errorf("Failed to parse request JSON: %s", err)
		return 500, ErrorResponse("Failed to parse request JSON.")
	}

	user := sess.User
	changed := 0
	name, _ := data["name"].(string)
	if name != "" {
		user.Name = name
		changed++
	}
	email, _ := data["email"].(string)
	if email != "" {
		user.Email = email
		changed++
	}

	if changed == 0 {
		return 204, "" // Nothing to update
	}

	// Update data
	err = repository.UserUpdate(sess, user)
	if err != nil {
		log.Errorf("Failed to update user for patching profile: %s", err)
		return 500, ErrorResponse("Failed to update user for patching profile.")
	}

	return 204, ""
}

func OptionsProfile(ctx *macaron.Context) (int, string) {
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Methods",
		"GET, PATCH, POST, OPTIONS",
	)
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Headers",
		"Content-Type, Authentication",
	)
	return 204, ""
}
