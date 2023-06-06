package app

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/urfave/cli/v2"
)

type App struct {
	ec2    *ec2.Client
	ssm    *ssm.Client
	region string
}

type appKey int

const (
	appCLI appKey = iota
)

func New(c *cli.Context) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	app := &App{
		ec2:    ec2.NewFromConfig(cfg),
		ssm:    ssm.NewFromConfig(cfg),
		region: cfg.Region,
	}

	c.Context = context.WithValue(c.Context, appCLI, app)

	return nil
}
