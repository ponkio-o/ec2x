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
	architecture    types.ArchitectureValues
	instanceType    types.InstanceType
	instanceID      string
	instanceProfile string
	keyName         string
	privateIP       string
	state           types.InstanceStateName
	nameTag         string
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
			return fmt.Sprintf("%s - %s (%s)", ins[i].instanceID, ins[i].nameTag, ins[i].privateIP)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n",
				"Name",
				ins[i].nameTag,
				"Architecture",
				ins[i].architecture,
				"InstanceType",
				ins[i].instanceType,
				"InstanceID",
				ins[i].instanceID,
				"InstanceProfile",
				ins[i].instanceProfile,
				"KeyName",
				ins[i].keyName,
				"PrivateIP",
				ins[i].privateIP,
				"State",
				ins[i].state,
			)
		},
		))
	if err != nil {
		log.Fatal(err)
	}
	return ins[idx].instanceID, nil
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
			instances = append(instances, EC2Instance{
				architecture:    rsv.Instances[0].Architecture,
				instanceType:    rsv.Instances[0].InstanceType,
				instanceID:      aws.ToString(rsv.Instances[0].InstanceId),
				instanceProfile: extractInstanceProfile(rsv.Instances[0].IamInstanceProfile),
				keyName:         extractKeyName(rsv.Instances[0].KeyName),
				privateIP:       aws.ToString(rsv.Instances[0].PrivateIpAddress),
				state:           rsv.Instances[0].State.Name,
				nameTag:         extractNameTag(rsv.Instances[0].Tags),
			})
		}
	}

	return instances, nil
}
