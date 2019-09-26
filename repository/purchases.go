package repository

import (
	"fmt"

	arango "github.com/arangodb/go-driver"
	"github.com/nu7hatch/gouuid"
)

const (
	COLLECTION_PURCHASES = "purchases"
)

type Purchase struct {
	Key string `json:"_key"`
	Category string `json:"category"`
	Venue string `json:"venue"`
	Date string `json:"date"`
	Month int `json:"month"`
	Year int `json:"year"`
	Sum string `json:"sum"`
}

func GetPurchases(month int, year int) (*[]Purchase, error) {
	db, err := GetDb(); if err != nil {
		return nil, err
	}

	// ----
	// Query database

	c, err := db.Query(
		ctx,
		"FOR p IN purchases FILTER p.month == @month AND p.year == @year RETURN p",
		map[string]interface{}{"month": month, "year": year},
	)
	defer c.Close()

	res := make([]Purchase, c.Count())
	for {
		var p Purchase
		_, err := c.ReadDocument(ctx, &p)

		if (arango.IsNoMoreDocuments(err)) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, p)
	}

	return &res, nil
}

func AddPurchase(purchase Purchase) (string, error) {
	col, err := GetCollection(COLLECTION_PURCHASES); if err != nil {
		return "", err
	}

	key, err := uuid.NewV4(); if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %s", err)
	}
	purchase.Key = key.String()

	_, err = col.CreateDocument(ctx, purchase); if err != nil {
		return "", err
	}

	return purchase.Key, nil
}

func UpdatePurchase(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_PURCHASES); if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeletePurchase(key string) error {
	col, err := GetCollection(COLLECTION_PURCHASES); if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}