package handler

import (
	"encoding/json"
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/repository"
)

func GetCategories(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	categories, err := repository.GetCategories()
	if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(categories)
}

func PutCategory(ctx *macaron.Context) (int, string) {
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
	// Extract name

	name, ok := data["name"].(string)
	if !ok {
		return 400, ErrorResponse("Parameter 'name' is required and must be a string")
	}

	//----
	// Create category

	category, err := repository.AddCategory(name)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to add category: %s", err))
	}

	return 200, SuccessResponse(category)
}

func PostCategory(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	key := ctx.Params(":key")
	if key == "" {
		return 400, ErrorResponse("No category key specified")
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

	// any data to update at all?
	if len(values) == 0 {
		return 200, SuccessResponse(nil)
	}

	// ----
	// Execute the update

	err = repository.UpdateCategory(key, &values)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to update category: %s", err))
	}

	return 200, SuccessResponse(nil)
}

func DeleteCategory(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	key := ctx.Params(":key")
	if key == "" {
		return 400, ErrorResponse("No category key specified")
	}

	err := repository.DeleteCategory(key)
	if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to delete category: %s", err))
	}
	return 200, SuccessResponse(nil)
}

func OptionsCategory(ctx *macaron.Context) (int, string) {
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
