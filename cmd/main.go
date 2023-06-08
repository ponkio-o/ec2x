package main

import (
	"fmt"
	"log"
	"os"

	app "github.com/ponkio-o/ec2x"
	"github.com/urfave/cli/v2"
)

var (
	version  = ""
	revision = ""
)

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("v%s (%s)\n", cCtx.App.Version, revision)
	}

	app := &cli.App{
		Name:    "ec2x",
		Usage:   "ec2x is connect to EC2 instance using SSM Session Manager",
		Version: version,
		Before:  app.New,
		Commands: []*cli.Command{
			{
				Name:   "connect",
				Usage:  "Connect to EC2 instance with Session Manager",
				Action: app.ConnectToEC2Instance,
			},
		},
		DefaultCommand: "connect",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
