package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/qeesung/asciiplayer/pkg/decoder"
	"github.com/qeesung/image2ascii/convert"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

var cinemaCmd = &cobra.Command{
	Use:   "cinema",
	Short: "Asciicinema of an ec2 console",
	Run:   cinemaCommandRun,
}

func init() {
	cinemaCmd.Flags().StringVar(&instanceId, "instance-id", "", "Instance id")
	cinemaCmd.Flags().StringVar(&region, "region", "", "Region definition")

	RootCmd.AddCommand(cinemaCmd)
}

func cinemaCommandRun(_ *cobra.Command, _ []string) {
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
	terminalPlayer := NewImageTerminalPlayer()

	for {
		rl.Take()

		ec2Screen, err := ec2Client.GetConsoleScreenshot(context.TODO(), &ec2.GetConsoleScreenshotInput{
			InstanceId: aws.String(instanceId),
		})
		if err != nil {
			log.Fatal(err)
		}

		terminalPlayer.Play(*ec2Screen.ImageData)
		time.Sleep(time.Microsecond)
		fmt.Fprint(os.Stdout, "\033[H\033[2J")
	}
}

type ImageTerminalPlayer struct {
	decoder   decoder.Decoder
	converter *convert.ImageConverter
}

// NewImageTerminalPlayer create a new ImageTerminalPlayer object
func NewImageTerminalPlayer() *ImageTerminalPlayer {
	return &ImageTerminalPlayer{
		decoder:   decoder.NewImageDecoder(),
		converter: convert.NewImageConverter(),
	}
}

func (player *ImageTerminalPlayer) Play(imagedata string) {
	// decode imagedata from base64 and convert to io.Reader
	data, err := base64.StdEncoding.DecodeString(imagedata)
	if err != nil {
		log.Fatal(err)
	}

	// decode the file first
	frames, err := player.decoder.Decode(bytes.NewReader(data), nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(frames) == 0 {
		log.Fatal("missing frames")
	}
	frame := frames[0]

	asciiImageStr := player.converter.Image2ASCIIString(frame, &convert.DefaultOptions)
	fmt.Fprint(os.Stdout, asciiImageStr)
}
