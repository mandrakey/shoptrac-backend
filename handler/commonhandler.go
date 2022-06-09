package handler

import (
	"github.com/mandrakey/shoptrac/config"
)

func GetVersion() (int, string) {
	return 200, SuccessResponse(map[string]interface{}{"version": config.AppVersion})
}
