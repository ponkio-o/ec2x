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

type sessInfo struct {
	SessionID  string
	StreamUrl  string
	TokenValue string
}

func ConnectToEC2Instance(c *cli.Context) error {
	id, err := selectInstance(c)
	if err != nil {
		return err
	}

	err = startSession(c, id)
	if err != nil {
		return err
	}

	return nil
}

func startSession(c *cli.Context, id string) error {
	app := c.Context.Value(appCLI).(*App)
	result, err := app.ssm.StartSession(context.TODO(), &ssm.StartSessionInput{
		Target: aws.String(id),
	})
	if err != nil {
		return err
	}

	sessi := sessInfo{
		SessionID:  aws.ToString(result.SessionId),
		StreamUrl:  aws.ToString(result.StreamUrl),
		TokenValue: aws.ToString(result.TokenValue),
	}

	sess, _ := json.Marshal(sessi)
	cmd := exec.Command("session-manager-plugin", string(sess), app.region, "StartSession")
	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func selectInstance(c *cli.Context) (string, error) {
	ins, err := getInstanceInfo(c)
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
			return fmt.Sprintf("%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n",
				"Name",
				ins[i].NameTag,
				"Architecture",
				ins[i].Architecture,
				"InstanceType",
				ins[i].InstanceType,
				"InstanceID",
				ins[i].InstanceID,
				"InstanceProfile",
				ins[i].InstanceProfile,
				"KeyName",
				ins[i].KeyName,
				"PrivateIP",
				ins[i].PrivateIP,
				"State",
				ins[i].State,
			)
		},
		))
	if err != nil {
		log.Fatal(err)
	}
	return ins[idx].InstanceID, nil
}

func getInstanceInfo(c *cli.Context) ([]EC2Instance, error) {
	app := c.Context.Value(appCLI).(*App)
	paginator := ec2.NewDescribeInstancesPaginator(app.ec2, &ec2.DescribeInstancesInput{
		MaxResults: aws.Int32(150),
	})

	var instances []EC2Instance
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, rsv := range page.Reservations {
			// Skip if instance is not running
			if rsv.Instances[0].State.Name != types.InstanceStateNameRunning {
				continue
			}
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
