package main

import (
	"log"
	"os"

	as "github.com/centrifuge/account-scraper"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "Centrifuge Chain Account Scraper",
		Description: "The scraper returns an scale encoded list of accountIDs file in out/accounts.scale",
		Usage: "requires URL of full archive node",
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name: "url",
				Value: "",
				Usage: "URL of full archive node",
			},
		},
		Action: func(c *cli.Context) error {
			return as.Process(c.String("url"))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
