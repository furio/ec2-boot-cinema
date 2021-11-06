package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

var singleLogCmd = &cobra.Command{
	Use:   "single-log",
	Short: "Logs of an ec2 console",
	Run:   singlelogCommandRun,
}

func init() {
	singleLogCmd.Flags().StringVar(&instanceId, "instance-id", "", "Instance id")
	singleLogCmd.Flags().StringVar(&region, "region", "", "Region definition")

	RootCmd.AddCommand(singleLogCmd)
}

func singlelogCommandRun(_ *cobra.Command, _ []string) {
	if instanceId == "" {
		log.Fatal("Please set the --instance-id option")
	}

	regionOpts := config.WithDefaultRegion("")
	if region != "" {
		regionOpts = config.WithRegion(region)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), regionOpts)
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	rl := ratelimit.New(15)

	for {
		rl.Take()

		ec2Log, err := ec2Client.GetConsoleOutput(context.TODO(), &ec2.GetConsoleOutputInput{
			InstanceId: aws.String(instanceId),
		})
		if err != nil {
			log.Fatal(err)
		}

		if ec2Log.Output != nil {
			data, err := base64.StdEncoding.DecodeString(*ec2Log.Output)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Fprint(os.Stdout, string(data))

			dataSplice := strings.Split(string(data), "\n")
			fmt.Fprint(os.Stdout, dataSplice[len(dataSplice)-3:])
			fmt.Fprint(os.Stdout, "|_|_|_|_|_|")
			fmt.Fprint(os.Stdout, dataSplice[len(dataSplice)-3])
			return
		}
	}
}
