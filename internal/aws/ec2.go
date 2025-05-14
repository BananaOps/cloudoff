package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Tag struct {
	Key   string
	Value string
}

type Instance struct {
	ID               string
	Name             string
	PrivateIpAddress string
	InstanceId       string
	Region           string
	State            string
	Tags             []Tag
}

func DiscoverEC2Instances() []Instance {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := ec2.NewFromConfig(cfg)

	filters := []types.Filter{
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running", "stopped"},
		},
	}

	// Define parameters to describe instances
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	// Request DescribeInstances
	result, err := svc.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to describe instances, %v", err)
	}

	var listInstances []Instance

	// for each instance in the result, get the name and the private ip address
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {

			// Get the name of the instance
			title := instance.InstanceId
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					title = tag.Value
					break
				}
			}

			listInstances = append(listInstances, Instance{
				ID:               *instance.InstanceId,
				Name:             *title,
				PrivateIpAddress: *instance.PrivateIpAddress,
				InstanceId:       *instance.InstanceId,
				Region:           svc.Options().Region,
				State:            string(instance.State.Name),
				Tags:             ConvertToCustomTag(instance.Tags),
			})
		}
	}

	return listInstances
}

// Fonction pour convertir une InstanceTag en CustomTag
func ConvertToCustomTag(instanceTag []types.Tag) []Tag {

	var customTags []Tag
	for _, tag := range instanceTag {
		customTags = append(customTags, Tag{
			Key:   *tag.Key,
			Value: *tag.Value,
		})
	}
	return customTags

}
