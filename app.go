package app

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hairyhenderson/go-which"
	"github.com/urfave/cli/v2"
)

type App struct {
	EC2    *ec2.Client
	SSM    *ssm.Client
	Region string
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

	// check session-manager-plugin
	if !which.Found("session-manager-plugin") {
		return errors.New("you need to install session-manager-plugin command")
	}

	app := &App{
		EC2:    ec2.NewFromConfig(cfg),
		SSM:    ssm.NewFromConfig(cfg),
		Region: cfg.Region,
	}

	c.Context = context.WithValue(c.Context, appCLI, app)

	return nil
}
