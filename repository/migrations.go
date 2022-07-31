package repository

import (
	"fmt"
	"time"

	arango "github.com/arangodb/go-driver"
	"github.com/mandrakey/shoptrac/config"
)

const COLLECTION_SHOPTRAC_MIGRATIONS = "shoptrac_migrations"

type Migration struct {
	Version int    `json:"version"`
	Created string `json:"created"`
}

func RunMigrations() error {
	log := config.Logger()
	log.Debug("RunMigrations()")

	db, col, err := prepare()
	if err != nil {
		return err
	}

	currentVersion, err := getCurrentDbVersion(db)
	if err != nil {
		return err
	}
	log.Infof("Current database version: %d", currentVersion)

	newVersion, err := migrate(db, col, currentVersion)
	if err != nil {
		return err
	}

	log.Infof("Finished migration to database version %d.", newVersion)
	return nil
}

func prepare() (arango.Database, arango.Collection, error) {
	db, err := GetDb()
	if err != nil {
		return nil, nil, err
	}

	exists, err := db.CollectionExists(ctx, COLLECTION_SHOPTRAC_MIGRATIONS)
	if err != nil {
		return nil, nil, err
	}

	var col arango.Collection
	if !exists {
		col, err = createMigrationsCollection(db)
		if err != nil {
			return nil, nil, err
		}
	} else {
		col, err = db.Collection(ctx, COLLECTION_SHOPTRAC_MIGRATIONS)
		if err != nil {
			return nil, nil, err
		}
	}

	return db, col, nil
}

func createMigrationsCollection(db arango.Database) (arango.Collection, error) {
	opts := arango.CreateCollectionOptions{
		KeyOptions: &arango.CollectionKeyOptions{
			AllowUserKeys: true,
		},
	}

	col, err := db.CreateCollection(ctx, COLLECTION_SHOPTRAC_MIGRATIONS, &opts)
	if err != nil {
		return nil, err
	}

	// Add initial migration
	m := Migration{Version: 1, Created: DateTimeToDb(time.Now().UTC())}
	_, err = col.CreateDocument(ctx, m)
	if err != nil {
		return nil, err
	}

	return col, nil
}

func getCurrentDbVersion(db arango.Database) (int, error) {
	c, err := db.Query(
		ctx,
		"FOR m IN shoptrac_migrations SORT m.version DESC LIMIT 1 RETURN m.version",
		nil,
	)
	if err != nil {
		return -1, err
	}
	defer c.Close()

	var version int
	_, err = c.ReadDocument(ctx, &version)
	if err != nil {
		return -1, fmt.Errorf("Failed to read current version: %s", err)
	}

	return version, nil
}

func addFinishedMigration(migrationsCollection arango.Collection, version int) error {
	m := Migration{Version: version, Created: DateTimeToDb(time.Now().UTC())}
	_, err := migrationsCollection.CreateDocument(ctx, m)
	return err
}

func migrate(db arango.Database, migrationsCollection arango.Collection, current int) (int, error) {
	if current < 1 {
		return -1, fmt.Errorf("Version can not be less than 1. Provided: %d", current)
	}

	log := config.Logger()

	finished := 1
	switch current {
	case 1:
		err := migrateFrom1(db)
		if err != nil {
			return finished, err
		}

		finished = 2
		err = addFinishedMigration(migrationsCollection, finished)
		if err != nil {
			log.Errorf("Failed to save finished migration information for version %d: %s", finished, err)
			return finished, err
		}
		break

	default:
		log.Infof("No migration from version %d.", current)
	}

	return finished, nil
}

func migrateFrom1(db arango.Database) error {
	log := config.Logger()
	log.Info("Starting database migration to version 2.")

	defaultShopperKey := "unknown"

	// Add collection for shoppers
	shoppersExists, err := db.CollectionExists(ctx, COLLECTION_SHOPPERS)
	if err != nil {
		return err
	}
	if !shoppersExists {
		log.Info("Creating collection for shoppers ...")
		opts := arango.CreateCollectionOptions{
			KeyOptions: &arango.CollectionKeyOptions{
				AllowUserKeys: true,
			},
		}
		shoppersCollection, err := db.CreateCollection(ctx, COLLECTION_SHOPPERS, &opts)
		if err != nil {
			return fmt.Errorf("Failed to create collection '%s': %s", COLLECTION_SHOPPERS, err)
		}

		// Add default shopper document
		s := Shopper{
			Key:   "1",
			Name:  "N/A",
			Image: "",
		}
		_, err = shoppersCollection.CreateDocument(ctx, s)
		if err != nil {
			return fmt.Errorf("Failed to create default unknown shopper: %s", err)
		}
	}

	// Add default shopper to all purchases
	log.Info("Adding default shopper to all purchases ...")

	purchasesCollection, err := db.Collection(ctx, COLLECTION_PURCHASES)
	if err != nil {
		return fmt.Errorf("Failed to access purchases collection: %s", err)
	}

	c, err := db.Query(
		ctx,
		"FOR p IN purchases RETURN { _key: p._key, shopper: p.shopper }",
		nil,
	)
	if err != nil {
		return fmt.Errorf("Failed to query purchases to add default shopper to: %s", err)
	}
	defer c.Close()

	updateData := map[string]string{"shopper": defaultShopperKey}
	p := map[string]string{"_key": "", "shopper": ""}
	for {
		_, err = c.ReadDocument(ctx, &p)

		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return fmt.Errorf("Failed to retrieve purchase entry: %s", err)
		}

		if p["shopper"] == "" {
			_, err = purchasesCollection.UpdateDocument(ctx, p["_key"], updateData)
			if err != nil {
				log.Warningf("Failed to add default shopper to purchase '%s': %s", p["_key"], err)
			}
		}
	}

	return nil
}
