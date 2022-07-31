package repository

import (
	"fmt"
	"strconv"

	arango "github.com/arangodb/go-driver"
)

const (
	COLLECTION_SHOPPERS = "shoppers"
)

type Shopper struct {
	Key   string `json:"_key"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func GetShoppers() (*[]Shopper, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	c, err := db.Query(ctx, "FOR s IN shoppers RETURN s", nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res := make([]Shopper, c.Count())
	for {
		var s Shopper
		_, err := c.ReadDocument(ctx, &s)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, s)
	}

	return &res, nil
}

func AddShopper(name string, image string) (*Shopper, error) {
	col, err := GetCollection(COLLECTION_SHOPPERS)
	if err != nil {
		return nil, err
	}

	maxId, err := getMaxShopperIdInt()
	if arango.IsNoMoreDocuments(err) {
		maxId = 0
	} else if err != nil {
		return nil, fmt.Errorf("failed to get current highest shopper id: %s", err)
	}

	s := Shopper{Key: fmt.Sprintf("%d", maxId+1), Name: name, Image: image}
	_, err = col.CreateDocument(ctx, s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func UpdateShopper(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_SHOPPERS)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeleteShopper(key string) error {
	col, err := GetCollection(COLLECTION_SHOPPERS)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

func getMaxShopperIdInt() (int, error) {
	db, err := GetDb()
	if err != nil {
		return -1, err
	}

	c, err := db.Query(ctx, "FOR s IN shoppers SORT s._key DESC LIMIT 1 RETURN s._key", nil)
	if err != nil {
		return -1, err
	}
	defer c.Close()

	var key string
	_, err = c.ReadDocument(ctx, &key)
	if err != nil {
		return -1, err
	}

	intkey, err := strconv.ParseInt(key, 10, 0)
	if err != nil {
		return -1, err
	}

	return int(intkey), nil
}
