package repository

import (
	"fmt"

	arango "github.com/arangodb/go-driver"
	uuid "github.com/nu7hatch/gouuid"
)

const (
	COLLECTION_PURCHASES = "purchases"
)

type Purchase struct {
	Key      string `json:"_key"`
	Category string `json:"category"`
	Venue    string `json:"venue"`
	Shopper  string `json:"shopper"`
	Date     string `json:"date"`
	Month    int    `json:"month"`
	Year     int    `json:"year"`
	Sum      string `json:"sum"`
}

type PurchaseTimestamp struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

func GetPurchases(month int, year int) (*[]Purchase, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	// ----
	// Query database

	c, err := db.Query(
		ctx,
		"FOR p IN purchases FILTER p.month == @month AND p.year == @year SORT p.date DESC RETURN p",
		map[string]interface{}{"month": month, "year": year},
	)
	defer c.Close()

	res := make([]Purchase, c.Count())
	for {
		var p Purchase
		_, err := c.ReadDocument(ctx, &p)

		if p.Sum == "" {
			p.Sum = "0.0"
		}

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	return &res, nil
}

func GetPurchaseTimestamps() (*[]PurchaseTimestamp, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(
		ctx,
		"FOR p IN purchases COLLECT dates = { month: p.month, year: p.year } SORT dates.year, dates.month RETURN dates",
		nil,
	)
	defer c.Close()

	res := make([]PurchaseTimestamp, c.Count())
	for {
		var t PurchaseTimestamp
		_, err := c.ReadDocument(ctx, &t)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, t)
	}

	return &res, nil
}

func GetPurchasesCountForShopper(shopperKey string) (int, error) {
	db, err := GetDb()
	if err != nil {
		return -1, err
	}

	c, err := db.Query(
		ctx,
		"FOR p IN purchases FILTER p.shopper == @shopperKey COLLECT WITH COUNT INTO cnt RETURN cnt",
		map[string]interface{}{"shopperKey": shopperKey},
	)
	if err != nil {
		return -1, fmt.Errorf("Failed to prepare query for counting purchases for given shopper: %s", err)
	}
	defer c.Close()

	var cnt int
	_, err = c.ReadDocument(ctx, &cnt)
	if err != nil {
		return -1, fmt.Errorf("Failed to read count from result set: %s", err)
	}

	return cnt, nil
}

func AddPurchase(purchase Purchase) (string, error) {
	col, err := GetCollection(COLLECTION_PURCHASES)
	if err != nil {
		return "", err
	}

	key, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %s", err)
	}
	purchase.Key = key.String()

	err = validatePurchase(&purchase)
	if err != nil {
		return "", err
	}

	_, err = col.CreateDocument(ctx, purchase)
	if err != nil {
		return "", err
	}

	return purchase.Key, nil
}

func UpdatePurchase(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_PURCHASES)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeletePurchase(key string) error {
	col, err := GetCollection(COLLECTION_PURCHASES)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

func validatePurchase(p *Purchase) error {
	if p.Key == "" {
		return fmt.Errorf("Missing purchase key")
	}
	if p.Category == "" {
		return fmt.Errorf("Missing purchase category")
	}
	if p.Venue == "" {
		return fmt.Errorf("Missing purchase venue")
	}
	if p.Shopper == "" {
		return fmt.Errorf("Missing purchase shopper")
	}
	if p.Date == "" {
		return fmt.Errorf("Missing purchase date")
	}
	if p.Month < 0 || p.Month > 12 {
		return fmt.Errorf("Invalid purchase month '%d'", p.Month)
	}
	if p.Sum == "" {
		return fmt.Errorf("Missing purchase sum")
	}

	return nil
}
