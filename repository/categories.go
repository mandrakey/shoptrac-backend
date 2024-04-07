/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package repository

import (
	"fmt"
	"strconv"

	arango "github.com/arangodb/go-driver"
)

const (
	COLLECTION_CATEGORIES = "categories"
)

type Category struct {
	Key  string `json:"_key"`
	Name string `json:"name"`
}

func GetCategories() (*[]Category, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	// ----
	// Query database

	c, err := db.Query(ctx, "FOR c IN categories RETURN c", nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res := make([]Category, c.Count())
	for {
		var cat Category
		_, err := c.ReadDocument(ctx, &cat)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, cat)
	}

	return &res, nil
}

func AddCategory(name string) (*Category, error) {
	col, err := GetCollection(COLLECTION_CATEGORIES)
	if err != nil {
		return nil, err
	}

	// Get new category id
	maxId, err := getMaxCategoryIdInt()
	if arango.IsNoMoreDocuments(err) {
		maxId = 0
	} else if err != nil {
		return nil, fmt.Errorf("failed to get highest current category id: %s", err)
	}

	// Create new Category and store
	cat := Category{Key: fmt.Sprintf("%d", maxId+1), Name: name}
	_, err = col.CreateDocument(ctx, cat)
	if err != nil {
		return nil, err
	}

	return &cat, nil
}

func UpdateCategory(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_CATEGORIES)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeleteCategory(key string) error {
	col, err := GetCollection(COLLECTION_CATEGORIES)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

func getMaxCategoryIdInt() (int, error) {
	db, err := GetDb()
	if err != nil {
		return -1, err
	}

	c, err := db.Query(ctx, "FOR c IN categories SORT c._key DESC LIMIT 1 RETURN c._key", nil)
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
