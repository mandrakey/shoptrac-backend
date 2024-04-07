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
	COLLECTION_VENUES = "venues"
)

type Venue struct {
	Key   string `json:"_key"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func GetVenues() (*[]Venue, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	// ----
	// Query database

	c, err := db.Query(ctx, "FOR v IN venues RETURN v", nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res := make([]Venue, c.Count())
	for {
		var v Venue
		_, err := c.ReadDocument(ctx, &v)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		res = append(res, v)
	}

	return &res, nil
}

func AddVenue(name string, image string) (*Venue, error) {
	col, err := GetCollection(COLLECTION_VENUES)
	if err != nil {
		return nil, err
	}

	// Get new venue id
	maxId, err := getMaxVenueIdInt()
	if arango.IsNoMoreDocuments(err) {
		maxId = 0
	} else if err != nil {
		return nil, fmt.Errorf("failed to get current highest venue id: %s", err)
	}

	// Create Venue and store
	v := Venue{Key: fmt.Sprintf("%d", maxId+1), Name: name, Image: image}
	_, err = col.CreateDocument(ctx, v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func UpdateVenue(key string, data *map[string]interface{}) error {
	col, err := GetCollection(COLLECTION_VENUES)
	if err != nil {
		return err
	}

	_, err = col.UpdateDocument(ctx, key, data)
	return err
}

func DeleteVenue(key string) error {
	col, err := GetCollection(COLLECTION_VENUES)
	if err != nil {
		return err
	}

	_, err = col.RemoveDocument(ctx, key)
	return err
}

func getMaxVenueIdInt() (int, error) {
	db, err := GetDb()
	if err != nil {
		return -1, err
	}

	c, err := db.Query(ctx, "FOR v IN venues SORT v._key DESC LIMIT 1 RETURN v._key", nil)
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
