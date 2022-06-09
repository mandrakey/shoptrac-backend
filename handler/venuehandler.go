package handler

import (
	"encoding/json"
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/repository"
)

func GetVenues(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	venues, err := repository.GetVenues()
	if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(venues)
}

func PutVenue(ctx *macaron.Context) (int, string) {
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
	// Extract name and image

	name, ok := data["name"].(string)
	if !ok {
		return 400, ErrorResponse("Parameter 'name' is required and must be a string")
	}

	image, _ := data["image"].(string)

	// ----
	// Create venue

	venue, err := repository.AddVenue(name, image)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to add venue: %s", err))
	}

	return 200, SuccessResponse(venue)
}

func PostVenue(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	key := ctx.Params(":key")
	if key == "" {
		return 400, ErrorResponse("No venue ket specified")
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
	// Extract updated data

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

	// any data to update at all?
	if len(values) == 0 {
		return 200, SuccessResponse(nil)
	}

	// ----
	// Execute the update

	err = repository.UpdateVenue(key, &values)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to update venue: %s", err))
	}

	return 200, SuccessResponse(nil)
}

func DeleteVenue(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	key := ctx.Params(":key")
	if key == "" {
		return 400, ErrorResponse("No venue key specified")
	}

	err := repository.DeleteVenue(key)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to delete venue: %s", err))
	}
	return 200, SuccessResponse(nil)
}

func OptionsVenue(ctx *macaron.Context) (int, string) {
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Methods",
		"GET, POST, PUT, DELETE, OPTIONS",
	)
	ctx.Resp.Header().Add(
		"Access-Control-Allow-Headers",
		"Content-Type, Authentication",
	)
	return 200, ""
}
