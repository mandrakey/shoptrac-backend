package repository

import (
	"fmt"
	"strconv"

	arango "github.com/arangodb/go-driver"
)

type CountSumHolder struct {
	Count int     `json:"count"`
	Sum   float64 `json:"sum"`
}

func GetOverviewStatistics(month int, year int) (map[string]*CountSumHolder, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	// ----
	// Query database

	qry := `LET currentMonth = (
		FOR p IN purchases
		FILTER p.month == @month AND p.year == @year
		COLLECT AGGREGATE sum = SUM(to_number(p.sum)), cnt = COUNT(p)
		RETURN { count: cnt, sum: sum != null ? sum : 0 }
	)
	LET lastMonth = (
		FOR p IN purchases
		FILTER p.month == @lastMonth AND p.year == @lastYear
		COLLECT AGGREGATE sum = SUM(to_number(p.sum)), cnt = COUNT(p)
		RETURN { count: cnt, sum: sum != null ? sum : 0 }
	)
	LET allTime = (
		FOR p IN purchases
		COLLECT AGGREGATE sum = SUM(to_number(p.sum)), cnt = COUNT(p)
		RETURN { count: cnt, sum: sum != null ? sum : 0 }
	)
	RETURN {
		lastMonth: lastMonth[0],
		currentMonth: currentMonth[0],
		allTime: allTime[0]
	}`

	lastMonth := month - 1
	lastYear := year
	if month < 1 {
		month = 12
		lastYear -= 1
	}

	data := map[string]interface{}{"month": month, "year": year, "lastMonth": lastMonth, "lastYear": lastYear}
	c, err := db.Query(ctx, qry, data)
	defer c.Close()

	// ----
	// Read results

	var res map[string]*CountSumHolder
	_, err = c.ReadDocument(ctx, &res)

	if err != nil && !arango.IsNoMoreDocuments(err) {
		return nil, err
	}

	// Round sums
	res["currentMonth"].Sum, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res["currentMonth"].Sum), 64)
	res["lastMonth"].Sum, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res["lastMonth"].Sum), 64)
	res["allTime"].Sum, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res["allTime"].Sum), 64)

	return res, nil
}

func GetPurchasesUnfiltered() (map[string]interface{}, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	//----
	// Query database

	qry := `LET years = (
		FOR p1 IN purchases
		COLLECT data = p1.year
		SORT data
		RETURN data
	)
	LET purchaselist = (
		FOR p IN purchases
		RETURN {
			"month": p.month,
			"year": p.year,
			"venue": p.venue,
			"category": p.category,
			"sum": p.sum
		}
	)
	RETURN {
		"meta": {
			"years": years
		},
		"purchases": purchaselist
	}`

	c, err := db.Query(ctx, qry, nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// ----
	// Read result

	qryResult := make(map[string]interface{})
	_, err = c.ReadDocument(ctx, &qryResult)
	if err != nil {
		return nil, err
	}

	return qryResult, nil
}
