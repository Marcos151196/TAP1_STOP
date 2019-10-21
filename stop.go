package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	aws "github.com/aws/aws-sdk-go/aws"
	session "github.com/aws/aws-sdk-go/aws/session"
	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	viper "github.com/spf13/viper"
)

var cfgFile string

var sess, _ = session.NewSession(&aws.Config{
	Region: aws.String("eu-west-2")},
)
var ec2svc = ec2.New(sess)

func main() {
	var err error
	var args1 string
	if len(os.Args) > 1 {
		args1 = os.Args[1]
	} else {
		args1 = "0"
	}
	if args1 == "1" {
		cfgFile = "config/configImproved.toml"
	} else {
		cfgFile = "config/config.toml"
	}

	// CONFIG FILE
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("[INIT] Unable to read config from file %s: %v\n", cfgFile, err)
		return
	} else {
		fmt.Printf("[INIT] Read configuration from file %s\n", cfgFile)
	}

	// stop MQTT BROKER
	args := []string{"-i", viper.GetString("general.sshcert"), "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", "ubuntu@" + viper.GetString("mqtt.dns"), "nohup sh " + viper.GetString("mqtt.stopscript") + " > /dev/null 2>&1 &"}
	output, err := RunCMD("ssh", args, true)
	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

	// stop EC2 ECHOSEARCH INSTANCES
	// Service discovery
	var IPList []string
	resp, err := ec2svc.DescribeInstances(nil)

	if err != nil {
		panic(err)
	}

	fmt.Println("> Number of reservation sets: ", len(resp.Reservations))
	for idx, res := range resp.Reservations {
		fmt.Println("  > Number of instances: ", len(res.Instances))

		for _, inst := range resp.Reservations[idx].Instances {
			for _, tag := range inst.Tags {
				if *tag.Value == "echosearch" {
					fmt.Println("    - Instance Tag Key: ", *tag.Key)
					fmt.Println("    - Instance Tag Value: ", *tag.Value)
					fmt.Println("    - Instance ID: ", *inst.InstanceId)
					fmt.Println("    - Instance PublicIPAddress: ", *inst.PublicDnsName)
					if *inst.PublicDnsName != "" {
						IPList = append(IPList, *inst.PublicDnsName)
					}
				}
			}
		}
	}

	// stop apps
	fmt.Println(IPList)
	for _, IP := range IPList {
		args := []string{"-i", viper.GetString("general.sshcert"), "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", "ubuntu@" + IP, "nohup sh " + viper.GetString("echosearch.stopscript") + " > /dev/null 2>&1 &"}
		output, err := RunCMD("ssh", args, true)
		if err != nil {
			fmt.Println("Error:", output)
		} else {
			fmt.Println("Result:", output)
		}
	}

	// stop WEBCLIENT
	args = []string{"-i", viper.GetString("general.sshcert"), "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", "ubuntu@" + viper.GetString("webclient.IP"), "nohup sh " + viper.GetString("webclient.stopscript") + " > /dev/null 2>&1 &"}
	output, err = RunCMD("ssh", args, true)
	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

}

// RunCMD is a simple wrapper around terminal commands
func RunCMD(path string, args []string, debug bool) (out string, err error) {
	cmd := exec.Command(path, args...)
	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)
	if debug {
		fmt.Println(strings.Join(cmd.Args[:], " "))
		if err != nil {
			fmt.Println("RunCMD ERROR")
			fmt.Println(out)
		}
	}
	return
}
