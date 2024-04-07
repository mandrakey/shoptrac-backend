/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/repository"
	"gopkg.in/macaron.v1"
)

func GetShoppers(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	shoppers, err := repository.GetShoppers()
	if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(shoppers)
}

func PutShoppers(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to read request: %s", err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to parse JSON: %s", err))
	}

	// ----
	// Extract data

	name, ok := data["name"].(string)
	if !ok {
		return 400, ErrorResponse("Parameter 'name' is required and must be a string")
	}
	image, _ := data["image"].(string)

	shopper, err := repository.AddShopper(name, image)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to add shopper: %s", err))
	}

	return 200, SuccessResponse(shopper)
}

func PatchShoppers(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	key := ctx.Params(":key")
	if key == "" {
		return 400, ErrorResponse("No shopper key specified")
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to read request: %s", err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to parse JSON: %s", err))
	}

	// ----
	// Extract data

	values := make(map[string]interface{})

	// name
	if data["name"] != nil {
		name, ok := data["name"].(string)
		if !ok {
			return 400, ErrorResponse("The parameter 'name' must be a string")
		}
		if name == "" {
			return 400, ErrorResponse("The parameter 'name' must not be empty")
		}
		values["name"] = name
	}

	// image
	if data["image"] != nil {
		image, ok := data["image"].(string)
		if !ok {
			return 400, ErrorResponse("The parameter 'image' must be a base64 encoded string")
		}
		values["image"] = image
	}

	if len(values) == 0 {
		return 200, SuccessResponse(nil)
	}

	// ----
	// Execute

	err = repository.UpdateShopper(key, &values)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to update shopper: %s", err))
	}

	return 200, SuccessResponse(nil)
}

func DeleteShoppers(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	log := config.Logger()

	key := ctx.Params("key")
	if key == "" {
		return 400, ErrorResponse("No shopper key specified")
	}

	purchasesCount, err := repository.GetPurchasesCountForShopper(key)
	if err != nil {
		log.Errorf("Failed to check usage count for shopper %s before deleting: %s", key, err)
		return 500, ErrorResponse(fmt.Sprintf("Failed to check usafe for shopper"))
	}

	if purchasesCount > 0 {
		return 400, ErrorResponse("Cannot delete a shopper currently in use.")
	}

	err = repository.DeleteShopper(key)
	if err != nil {
		log.Errorf("Failed to delete shopper: %s", err)
		return 500, ErrorResponse("Failed to delete shopper.")
	}
	return 200, SuccessResponse(nil)
}

func OptionsShoppers(ctx *macaron.Context) (int, string) {
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Methods",
		"GET, PATCH, PUT, DELETE, OPTIONS",
	)
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Headers",
		"Content-Type, Authentication",
	)
	return 200, ""
}
