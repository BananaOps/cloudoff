package ec2

import (
	"context"
	"fmt"
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
	Spot             bool
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

			var spot bool = false

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

func StopInstance(instanceID, region string) error {
	// Charger la configuration AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("erreur lors du chargement de la configuration AWS : %v", err)
	}

	// Créer un client EC2
	ec2Client := ec2.NewFromConfig(cfg)

	// Préparer l'entrée pour StopInstances
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Appeler StopInstances
	_, err = ec2Client.StopInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("erreur lors de l'arrêt de l'instance %s : %v", instanceID, err)
	}

	fmt.Printf("Instance %s arrêtée avec succès.\n", instanceID)
	return nil
}

func StartInstance(instanceID, region string) error {
	// Charger la configuration AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("erreur lors du chargement de la configuration AWS : %v", err)
	}

	// Créer un client EC2
	ec2Client := ec2.NewFromConfig(cfg)

	// Préparer l'entrée pour StopInstances
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Appeler StopInstances
	_, err = ec2Client.StartInstances(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("erreur lors du démarrage de l'instance %s : %v", instanceID, err)
	}

	fmt.Printf("Instance %s démarré avec succès.\n", instanceID)
	return nil
}
