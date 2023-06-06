package main

import (
	"log"
	"os"

	app "github.com/ponkio-o/ec2x"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "ec2x",
		Description: "ec2x is connect to EC2 instance using SSM Session Manager",
		Before:      app.New,
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
