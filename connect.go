package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmt "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v2"
)

type EC2Instance struct {
	Architecture    types.ArchitectureValues
	InstanceType    types.InstanceType
	InstanceID      string
	InstanceProfile string
	KeyName         string
	PrivateIP       string
	State           types.InstanceStateName
	NameTag         string
}

func ConnectCommand(c *cli.Context) error {
	app := c.Context.Value(appCLI).(*App)
	id, err := app.selectInstance()
	if err != nil {
		return err
	}

	err = app.startSession(id, app.Region)
	if err != nil {
		return err
	}

	return nil
}

func (app App) startSession(id, region string) error {
	result, err := app.SSM.StartSession(context.TODO(), &ssm.StartSessionInput{
		Target: aws.String(id),
	})
	if err != nil {
		return err
	}

	sess, err := json.Marshal(result)
	if err != nil {
		return err
	}
	cmd := exec.Command("session-manager-plugin", string(sess), region, "StartSession")
	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (app App) listSSMManagedInstances() ([]string, error) {
	paginator := ssm.NewDescribeInstanceInformationPaginator(app.SSM, &ssm.DescribeInstanceInformationInput{
		Filters: []ssmt.InstanceInformationStringFilter{
			{
				Key:    aws.String("PingStatus"),
				Values: []string{"Online"},
			},
		},
	})

	var managedInstances []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		for _, ins := range page.InstanceInformationList {
			managedInstances = append(managedInstances, aws.ToString(ins.InstanceId))
		}
	}

	return managedInstances, nil
}

func (app App) selectInstance() (string, error) {
	managedInstances, err := app.listSSMManagedInstances()
	if err != nil {
		return "", err
	}

	ins, err := app.getInstanceInfo(managedInstances)
	if err != nil {
		return "", err
	}
	idx, err := fuzzyfinder.Find(
		ins,
		func(i int) string {
			return fmt.Sprintf("%s - %s (%s)", ins[i].InstanceID, ins[i].NameTag, ins[i].PrivateIP)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return genPreviewWindow(ins[i])
		},
		))
	if err != nil {
		log.Fatal(err)
	}
	return ins[idx].InstanceID, nil
}

func (app App) getInstanceInfo(managedInstances []string) ([]EC2Instance, error) {
	paginator := ec2.NewDescribeInstancesPaginator(app.EC2, &ec2.DescribeInstancesInput{
		InstanceIds: managedInstances,
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{string(types.InstanceStateNameRunning)},
			},
		},
	})

	var instances []EC2Instance
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, rsv := range page.Reservations {
			instances = append(instances, EC2Instance{
				Architecture:    rsv.Instances[0].Architecture,
				InstanceType:    rsv.Instances[0].InstanceType,
				InstanceID:      aws.ToString(rsv.Instances[0].InstanceId),
				InstanceProfile: extractInstanceProfile(rsv.Instances[0].IamInstanceProfile),
				KeyName:         extractKeyName(rsv.Instances[0].KeyName),
				PrivateIP:       aws.ToString(rsv.Instances[0].PrivateIpAddress),
				State:           rsv.Instances[0].State.Name,
				NameTag:         extractNameTag(rsv.Instances[0].Tags),
			})
		}
	}

	return instances, nil
}
