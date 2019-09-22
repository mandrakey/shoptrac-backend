package handler

import (
	"encoding/json"

	"github.com/mandrakey/shoptrac/config"
)

func ErrorResponse(message string) string {
	return ErrorResponseWithData(message, nil)
}

func ErrorResponseWithData(message string, data interface{}) string {
	obj := map[string]interface{}{
		"message": message,
		"data": data,
	}

	res, err := json.Marshal(obj); if err != nil {
		config.Logger().Errorf("Failes to generate error response: %s", err)
		return ""
	}

	return string(res)
}

func SuccessResponse(data interface{}) string {
	obj := map[string]interface{}{
		"data": data,
	}

	res, err := json.Marshal(obj); if err != nil {
		config.Logger().Errorf("Failes to generate success response: %s", err)
		return ""
	}

	return string(res)
}