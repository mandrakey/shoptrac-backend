package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"

	"github.com/mandrakey/shoptrac/config"
)

const (
	DATE_FORMAT     = "2006-01-02"
	DATETIME_FORMAT = "2006-01-02 15:04:05"
)

var (
	client arango.Client
	ctx    = context.Background()
)

func GetDb() (arango.Database, error) {
	cfg := config.GetAppConfig()

	if client == nil {
		conn, err := http.NewConnection(http.ConnectionConfig{
			Endpoints: []string{
				fmt.Sprintf("%s://%s:%d", cfg.Database.Protocol, cfg.Database.Host, cfg.Database.Port),
			},
		})
		if err != nil {
			return nil, err
		}

		client, err = arango.NewClient(arango.ClientConfig{
			Connection:     conn,
			Authentication: arango.BasicAuthentication(cfg.Database.User, cfg.Database.Password),
		})
		if err != nil {
			return nil, err
		}
	}

	db, err := client.Database(ctx, cfg.Database.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to access database %s: %s", cfg.Database.DatabaseName, err)
	}
	return db, nil
}

func GetCollection(name string) (arango.Collection, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}

	col, err := db.Collection(ctx, name)
	if err != nil {
		return nil, err
	}

	return col, nil
}

func BuildFilterString(filters map[string]interface{}, prefix string) string {
	f := make([]string, 0)
	for field, rawValue := range filters {
		value := rawValue.(string)

		// determine sign to use
		frst := string(value[0])
		sign := "=="
		if frst == "!" {
			sign = "!="
			value = value[1:len(value)]
		}
		if frst == "<" || frst == ">" {
			scnd := string(value[1])
			if scnd == "=" {
				sign = fmt.Sprintf("%s%s", frst, scnd)
				value = strings.Trim(value[2:len(value)], " ")
			} else {
				sign = frst
				value = strings.Trim(value[1:len(value)], " ")
			}
		}
		filters[field] = value

		// determine value type
		if value == "null" {
			f = append(f, fmt.Sprintf("%s%s %s null", prefix, field, sign))
			continue
		}
		vint, err := strconv.ParseInt(value, 0, 0)
		if err == nil {
			f = append(f, fmt.Sprintf("%s%s %s @%s", prefix, field, sign, field))
			filters[field] = vint
			continue
		}
		vfloat, err := strconv.ParseFloat(value, 32)
		if err == nil {
			f = append(f, fmt.Sprintf("%s%s %s @%s", prefix, field, sign, field))
			filters[field] = vfloat
			continue
		}

		//----
		// string value

		// wildcard?
		if strings.ContainsAny(value, "%_") {
			sign = "LIKE"
		}
		f = append(f, fmt.Sprintf("%s%s %s @%s", prefix, field, sign, field))
	}

	return strings.Join(f, " AND ")
}

func DateFromDb(v string) (time.Time, error) {
	return time.Parse(DATE_FORMAT, v)
}

func DateToDb(v time.Time) string {
	return v.Format(DATE_FORMAT)
}

func DateTimeFromDb(v string) (time.Time, error) {
	return time.Parse(DATETIME_FORMAT, v)
}

func DateTimeToDb(v time.Time) string {
	return v.Format(DATETIME_FORMAT)
}
