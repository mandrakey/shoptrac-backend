package handler

import (
	"fmt"
	"strconv"
	"time"
	"encoding/json"

	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/repository"
)

func GetPurchases(ctx *macaron.Context) (int, string) {
	// ----
	// Get month and year parameters

	pmonth := ctx.Params(":month")
	if pmonth == "" {
		return 400, ErrorResponse("Parameter 'month' is required and must be a number")
	}
	month, err := strconv.ParseInt(pmonth, 10, 0); if err != nil {
		return 400, ErrorResponse(fmt.Sprintf("Failed to parse month value: %s", err))
	}

	pyear := ctx.Params(":year")
	if pyear == "" {
		return 400, ErrorResponse("Parameter 'year' is required and must be a number")
	}
	year, err := strconv.ParseInt(pyear, 10, 0); if err != nil {
		return 400, ErrorResponse(fmt.Sprintf("Failed to parse year value: %s", err))
	}

	// ----
	// Get purchases

	purchases, err := repository.GetPurchases(int(month), int(year)); if err != nil {
		return 500, ErrorResponse(err.Error())
	}

	return 200, SuccessResponse(purchases)
}

func PutPurchase(ctx *macaron.Context) (int, string) {
	body, err := ctx.Req.Body().Bytes(); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to read request: %s", err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to parse JSON: %s", err))
	}

	// ----
	// Extract data

	purchase := repository.Purchase{}

	category, ok := data["category"].(int); if !ok {
		return 400, ErrorResponse("Parameter 'category' is required and must be a number")
	}
	purchase.Category = category

	venue, ok := data["venue"].(int); if !ok {
		return 400, ErrorResponse("Parameter 'venue' is required and must be a number")
	}
	purchase.Venue = venue

	date, ok := data["date"].(string); if !ok {
		return 400, ErrorResponse("Parameter 'date' is required and must be a string")
	}
	_, err = time.Parse(repository.DATE_FORMAT, date); if err != nil {
		return 400, ErrorResponse(fmt.Sprintf("Date '%s' is not valid: %s", date, err))
	}
	purchase.Date = date

	month, ok := data["month"].(int); if !ok {
		return 400, ErrorResponse("Parameter 'month' is required and must be a number")
	}
	purchase.Month = month

	year, ok := data["year"].(int); if !ok {
		return 400, ErrorResponse("Parameter 'year' is required and must be a number")
	}
	purchase.Year = year

	sum, ok := data["sum"].(string); if !ok {
		return 400, ErrorResponse("Parameter 'sum' is required and must be a string")
	}
	purchase.Sum = sum

	// ----
	// Create purchase

	key, err := repository.AddPurchase(purchase); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to add purchase: %s", err))
	}
	purchase.Key = key

	return 200, SuccessResponse(purchase)
}

func PostPurchase(ctx *macaron.Context) (int, string) {
	key := ctx.Params(":key"); if key == "" {
		return 400, ErrorResponse("No purchase ket specified")
	}

	body, err := ctx.Req.Body().Bytes(); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to read request: %s", err))
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to parse JSON: %s", err))
	}

	// ----
	// Extract updated data

	values := make(map[string]interface{})

	// category
	if data["category"] != nil {
		category, ok := data["category"].(int); if !ok {
			return 400, ErrorResponse("The parameter 'category' must be a number")
		}
		values["category"] = category
	}

	// venue
	if data["venue"] != nil {
		venue, ok := data["venue"].(int); if !ok {
			return 400, ErrorResponse("The parameter 'venue' must be a number")
		}
		values["venue"] = venue
	}

	// date
	if data["date"] != nil {
		date, ok := data["date"].(string); if !ok {
			return 400, ErrorResponse("The parameter 'date' must be a string")
		}
		_, err = time.Parse(repository.DATE_FORMAT, date); if err != nil {
			return 400, ErrorResponse(fmt.Sprintf("Date '%s' is not valid: %s", date, err))
		}
		values["date"] = date
	}
	
	// month
	if data["month"] != nil {
		month, ok := data["month"].(int); if !ok {
			return 400, ErrorResponse("The parameter 'month' must be a number")
		}
		values["month"] = month
	}

	// year
	if data["year"] != nil {
		year, ok := data["year"].(int); if !ok {
			return 400, ErrorResponse("The parameter 'year' must be a number")
		}
		values["year"] = year
	}

	// sum
	if data["sum"] != nil {
		sum, ok := data["sum"].(string); if !ok {
			return 400, ErrorResponse("The parameter 'sum' must be a string")
		}
		values["sum"] = sum
	}

	// any data to update at all?
	if len(values) == 0 {
		return 200, SuccessResponse(nil)
	}

	// ----
	// Execute the update

	err = repository.UpdatePurchase(key, &values); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to update purchase: %s", err))
	}

	return 200, SuccessResponse(nil)
}

func DeletePurchase(ctx *macaron.Context) (int, string) {
	key := ctx.Params(":key"); if key == "" {
		return 400, ErrorResponse("No purchase key specified")
	}

	err := repository.DeletePurchase(key); if err != nil {
		return 500, ErrorResponse(fmt.Sprintf("Failed to delete purchase: %s", err))
	}
	return 200, SuccessResponse(nil)
}