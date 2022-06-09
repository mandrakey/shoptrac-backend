package handler

import (
	"fmt"
	"strconv"

	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/repository"
)

func GetOverviewStatistics(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	// ----
	// Get month and year parameters

	pmonth := ctx.Params(":month")
	if pmonth == "" {
		return 400, ErrorResponse("Parameter 'month' is required and must be a number")
	}
	month, err := strconv.ParseInt(pmonth, 10, 0)
	if err != nil {
		return 400, ErrorResponse(fmt.Sprintf("Failed to parse month value: %s", err))
	}

	pyear := ctx.Params(":year")
	if pyear == "" {
		return 400, ErrorResponse("Parameter 'year' is required and must be a number")
	}
	year, err := strconv.ParseInt(pyear, 10, 0)
	if err != nil {
		return 400, ErrorResponse(fmt.Sprintf("Failed to parse year value: %s", err))
	}

	// ----
	// Get data

	stats, err := repository.GetOverviewStatistics(int(month), int(year))
	if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(stats)
}

func GetPurchasesUnfiltered(ctx *macaron.Context) (int, string) {
	if !IsValidSession(ctx) {
		return UnauthorizedResponse()
	}

	stats, err := repository.GetPurchasesUnfiltered()
	if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(stats)
}

func OptionsStatistics(ctx *macaron.Context) (int, string) {
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
