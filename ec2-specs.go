package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"net/url"
)

func main() {
	sess := session.New(&aws.Config{
		Region:      aws.String("ap-southeast-1"),
		Credentials: credentials.NewSharedCredentials("", "wls-prod-readonly"),
	})

	credential, err := sess.Config.Credentials.Get()

	if err != nil {
		fmt.Println("Error getting AWS credential")
		panic(err.Error())
	}

	fmt.Println("AWS_ACCESS_KEY_ID:", credential.AccessKeyID)
	fmt.Println("AWS_SECRET_ACCESS_KEY:", credential.SecretAccessKey)

	svc := ec2.New(sess)

	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("tag:Environment"),
			Values: []*string{aws.String("ctrl")},
		},
		&ec2.Filter{
			Name:   aws.String("tag:AWSComponent"),
			Values: []*string{aws.String("ec2")},
		},
	}

	params := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	results, err := svc.DescribeInstances(params)

	if err != nil {
		fmt.Println("Cannot list instances", err)
		return
	}

	fmt.Println("We have", len(results.Reservations), "instances")

	for index, _ := range results.Reservations {
		for _, instance := range results.Reservations[index].Instances {

			name := "None"
			for _, keys := range instance.Tags {
				if *keys.Key == "Name" {
					name = url.QueryEscape(*keys.Value)
				}
			}

			publicIp := ""

			if instance.PublicIpAddress != nil {
				publicIp = *instance.PublicIpAddress
			}

			fmt.Printf("%v, %v, %v\n", name, *instance.InstanceType, publicIp)
		}
	}
}
