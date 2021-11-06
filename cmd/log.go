package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Logs of an ec2 console",
	Run:   logCommandRun,
}

func init() {
	logCmd.Flags().StringVar(&instanceId, "instance-id", "", "Instance id")
	logCmd.Flags().StringVar(&region, "region", "", "Region definition")

	RootCmd.AddCommand(logCmd)
}

func logCommandRun(_ *cobra.Command, _ []string) {
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

			fmt.Fprint(os.Stdout, string(data))
			time.Sleep(time.Second)
			fmt.Fprint(os.Stdout, "\033[H\033[2J")
		}
	}
}
