package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"net/url"
	"strings"
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
			serviceName := ""
			serviceComponent := ""

			for _, keys := range instance.Tags {
				if *keys.Key == "Name" {
					name = url.QueryEscape(*keys.Value)
				}

				if *keys.Key == "Service" {
					serviceName = url.QueryEscape(*keys.Value)
				}

				if *keys.Key == "SeviceComponent" {
					serviceComponent = url.QueryEscape(*keys.Value)
				}
			}

			publicIp := "N/A"

			if instance.PublicIpAddress != nil {
				publicIp = *instance.PublicIpAddress
			}

			privateIp := *instance.PrivateIpAddress

			secGroup := []string{}

			for index, _ := range instance.SecurityGroups {
				secGroup = append(secGroup, *instance.SecurityGroups[index].GroupName)
			}

			ebsVolumns := []int64{}

			for index, _ := range instance.BlockDeviceMappings {
				params := &ec2.DescribeVolumesInput{
					VolumeIds: []*string{
						instance.BlockDeviceMappings[index].Ebs.VolumeId,
					},
				}

				resp, err := svc.DescribeVolumes(params)

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				for index, _ := range resp.Volumes {
					ebsVolumns = append(ebsVolumns, *resp.Volumes[index].Size)
				}
			}

			subnetParams := &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{
					instance.SubnetId,
				},
			}

			resp, err := svc.DescribeSubnets(subnetParams)

			if err != nil {
				fmt.Println(err.Error())
				return
			}

			subnets := []string{}

			for index, _ := range resp.Subnets {
				for _, keys := range resp.Subnets[index].Tags {
					if *keys.Key == "Name" {
						subnets = append(subnets, url.QueryEscape(*keys.Value))
					}
				}
			}

			instanceProfile := strings.Split(*instance.IamInstanceProfile.Arn, "/")[1]

			fmt.Printf("%v,%v,%v,%v,%v,%v,%v,%vGiB,%v,%v\n", name, serviceName, serviceComponent, *instance.InstanceType, publicIp, privateIp, subnets, ebsVolumns, secGroup, instanceProfile)
		}
	}
}
