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
			&cli.BoolFlag{
				Name: "append",
				Usage: "Appends to existing scale encoded file removing duplicates",
			},
		},
		Action: func(c *cli.Context) error {
			return as.Process(c.String("url"), c.Bool("append"))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
