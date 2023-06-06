package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v2"
)

type EC2Instance struct {
	architecture       string
	instanceType       string
	instanceID         string
	instanceProfileArn string
	keyName            string
	privateIP          string
	state              string
	nameTag            string
}

type sessInfo struct {
	SessionID  string
	StreamUrl  string
	TokenValue string
}

var instances = []EC2Instance{}

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
				func(p string) string {
					if p == "" {
						return "<None>"
					}
					return strings.Split(ins[i].instanceProfileArn, "/")[1]
				}(ins[i].instanceProfileArn),
				"KeyName",
				func(k string) string {
					if k == "" {
						return "<None>"
					}
					return k
				}(ins[i].keyName),
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
	result, err := app.ec2.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		MaxResults: aws.Int32(150),
	})
	if err != nil {
		return nil, err
	}

	for _, v := range result.Reservations {
		if aws.ToString((*string)(&v.Instances[0].State.Name)) == "running" {
			instances = append(instances, EC2Instance{
				architecture: aws.ToString((*string)(&v.Instances[0].Architecture)),
				instanceType: aws.ToString((*string)(&v.Instances[0].InstanceType)),
				instanceID:   aws.ToString(v.Instances[0].InstanceId),
				instanceProfileArn: func(p types.Instance) string {
					if p.IamInstanceProfile == nil {
						return ""
					}
					return aws.ToString(p.IamInstanceProfile.Arn)
				}(v.Instances[0]),
				keyName:   aws.ToString(v.Instances[0].KeyName),
				privateIP: aws.ToString(v.Instances[0].PrivateIpAddress),
				state:     aws.ToString((*string)(&v.Instances[0].State.Name)),
				nameTag: func(t []types.Tag) string {
					for _, v := range t {
						if aws.ToString(v.Key) == "Name" {
							return aws.ToString(v.Value)
						}
					}
					return ""
				}(v.Instances[0].Tags),
			})
		}
	}

	return instances, nil
}
