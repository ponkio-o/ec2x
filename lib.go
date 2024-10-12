package app

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func extractInstanceProfile(p *types.IamInstanceProfile) string {
	if p != nil {
		return strings.Split(aws.ToString(p.Arn), "/")[1]
	}

	return "<No instance profile>"
}

func extractKeyName(k *string) string {
	if aws.ToString(k) != "" {
		return aws.ToString(k)
	}

	return "<No key name>"
}

func extractNameTag(tags []types.Tag) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == "Name" {
			return aws.ToString(t.Value)
		}
	}

	return "<No Name Tag>"
}

func genPreviewWindow(ins EC2Instance) string {
	return fmt.Sprintf("%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n%-16s: %s\n",
		"Name", ins.NameTag,
		"Architecture", ins.Architecture,
		"InstanceType", ins.InstanceType,
		"InstanceID", ins.InstanceID,
		"InstanceProfile", ins.InstanceProfile,
		"KeyName", ins.KeyName,
		"PrivateIP", ins.PrivateIP,
		"State", ins.State,
	)
}
