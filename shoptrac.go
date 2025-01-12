/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli"
	"gopkg.in/macaron.v1"

	"github.com/mandrakey/shoptrac/config"
	"github.com/mandrakey/shoptrac/handler"
	"github.com/mandrakey/shoptrac/middleware"
	"github.com/mandrakey/shoptrac/repository"
)

var (
	configFile = "./shoptrac.json"
)

func main() {
	app := cli.NewApp()
	app.Name = "shoptrac - Shopping Tracker API"
	app.Usage = "Run the backend API"
	app.Version = config.AppVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "./shoptrac.json",
			Usage: "Load configuration from FILE",
		},
	}

	// Setup commands
	app.Commands = []cli.Command{
		{
			Name:   "serve",
			Usage:  "run the shoptrac API server",
			Action: runServe,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port",
					Usage: "Run on PORT",
				},
				cli.StringFlag{
					Name:  "address",
					Usage: "Run on ADDRESS",
				},
			},
		},
	}

	// Run
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
	}
}

func runServe(ctx *cli.Context) error {
	configFile := ctx.GlobalString("config")

	cfg := config.GetAppConfig()
	err := cfg.LoadFromFile(configFile)
	if err != nil {
		return fmt.Errorf("Failed to load configuration: %s", err)
	}
	config.SetupLogging(cfg.Logfile, cfg.Loglevel)
	log := config.Logger()

	// Run migrations
	err = repository.RunMigrations()
	if err != nil {
		err = fmt.Errorf("Failed to migrate database: %s", err)
		log.Error(err)
		return err
	}

	// Create server and set routing
	m := macaron.Classic()
	m.Use(config.IpFilterer(cfg))
	m.Use(middleware.SessionMiddleware())
	m.Use(func(ctx *macaron.Context) {
		ctx.Resp.Header().Add("Access-Control-Allow-Origin", cfg.CorsOrigin)
	})

	log.Infof("Starting shoptrac API server at %s:%d", cfg.Address, cfg.Port)

	m.Group("/api", func() {
		m.Group("/venues", func() {
			m.Get("/", handler.GetVenues)
			m.Put("/", handler.PutVenue)
			m.Post("/:key", handler.PostVenue)
			m.Delete("/:key", handler.DeleteVenue)

			m.Options("/", handler.OptionsVenue)
			m.Options("/*", handler.OptionsVenue)
		})
		m.Group("/categories", func() {
			m.Get("/", handler.GetCategories)
			m.Put("/", handler.PutCategory)
			m.Post("/:key", handler.PostCategory)
			m.Delete("/:key", handler.DeleteCategory)

			m.Options("/", handler.OptionsCategory)
			m.Options("/*", handler.OptionsCategory)
		})
		m.Group("/shoppers", func() {
			m.Get("/", handler.GetShoppers)
			m.Put("/", handler.PutShoppers)
			m.Patch("/:key", handler.PatchShoppers)
			m.Delete("/:key", handler.DeleteShoppers)

			m.Options("/", handler.OptionsShoppers)
			m.Options("/*", handler.OptionsShoppers)
		})
		m.Group("/purchases", func() {
			m.Get("/:year(\\d{4})/:month(\\d{1,2})", handler.GetPurchases)
			m.Put("/", handler.PutPurchase)
			m.Post("/:key", handler.PostPurchase)
			m.Delete("/:key", handler.DeletePurchase)

			m.Get("/timestamps", handler.GetPurchaseTimestamps)

			m.Options("/", handler.OptionsPurchase)
			m.Options("/*", handler.OptionsPurchase)
		})
		m.Group("/statistics", func() {
			m.Get("/overview/:year(\\d{4})/:month(\\d{1,2})", handler.GetOverviewStatistics)
			m.Get("/purchases_unfiltered", handler.GetPurchasesUnfiltered)
			m.Options("/*", handler.OptionsStatistics)
		})
		m.Group("/auth", func() {
			m.Get("/logout", handler.GetLogout)
			m.Post("/login", handler.PostLogin)
			m.Post("/continue", handler.PostContinue)
			m.Options("/*", handler.OptionsAuth)
		})
		m.Group("/users", func() {
			m.Get("/", handler.UsersGet)
			m.Get("/:uuid", handler.UsersGetUser)
			m.Put("/", handler.UsersPut)
			m.Patch("/", handler.UsersPatch)
			m.Delete("/:uuid", handler.UsersDelete)
			m.Options("/", handler.UsersOptions)
			m.Options("/*", handler.UsersOptions)
		})
		m.Group("/profile", func() {
			m.Get("/", handler.GetProfile)
			m.Patch("/", handler.PatchProfile)
			m.Post("/updatePassword", handler.PostProfileUpdatePassword)
			m.Options("/", handler.OptionsProfile)
			m.Options("/*", handler.OptionsProfile)
		})
		m.Get("/version", handler.GetVersion)
	})

	listenAddress := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	http.ListenAndServe(listenAddress, m)

	return nil
}
