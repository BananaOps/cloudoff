package ec2

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var logger *slog.Logger

type Tag struct {
	Key   string
	Value string
}

type Instance struct {
	Spot             bool
	ID               string
	Name             string
	PrivateIpAddress string
	InstanceId       string
	Region           string
	State            string
	LaunchTime       time.Time
	AttachTime       time.Time
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

	// For each instance in the result, get the name and the private IP address
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

			var spot = false

			if instance.InstanceLifecycle == "spot" {
				spot = true
			}

			listInstances = append(listInstances, Instance{
				Spot:             spot,
				ID:               *instance.InstanceId,
				Name:             *title,
				PrivateIpAddress: *instance.PrivateIpAddress,
				InstanceId:       *instance.InstanceId,
				Region:           svc.Options().Region,
				State:            string(instance.State.Name),
				Tags:             ConvertToCustomTag(instance.Tags),
				LaunchTime:       *instance.LaunchTime,
				AttachTime:       *instance.NetworkInterfaces[0].Attachment.AttachTime,
			})
		}
	}

	return listInstances
}

// Function to convert InstanceTag to CustomTag
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

func StopInstance(instanceID, region string) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %v", err)
	}

	// Create an EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Prepare input for StopInstances
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call StopInstances
	_, err = ec2Client.StopInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error stopping instance %s: %v", instanceID, err)
	}

	logger.Info("instance stopped successfully", "instance", instanceID)
	return nil
}

func StartInstance(instanceID, region string) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %v", err)
	}

	// Create an EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Prepare input for StartInstances
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call StartInstances
	_, err = ec2Client.StartInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error starting instance %s: %v", instanceID, err)
	}

	logger.Info("instance started successfully", "instance", instanceID)
	return nil
}

func TerminateInstance(instanceID, region string) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %v", err)
	}

	// Create an EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Prepare input for TerminateInstances
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call TerminateInstances
	_, err = ec2Client.TerminateInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error terminating instance %s: %v", instanceID, err)
	}

	logger.Info("instance terminated successfully", "instance", instanceID)
	return nil
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
