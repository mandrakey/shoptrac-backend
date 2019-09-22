package handler

import (
	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/config"
)

const (
	SESSION_KEY = "key"
)

func GetVersion(ctx *macaron.Context) (int, string) {
	return 200, SuccessResponse(map[string]interface{}{"version": config.AppVersion})
}