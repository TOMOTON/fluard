package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/spf13/pflag"
)

func main() {
	// Command-line flags
	tag := pflag.StringP("tag", "t", "fluard.test", "Fluentd event tag")
	recordInput := pflag.StringP("record", "r", "", "Event record as JSON string or @<file>")

	pflag.Parse()

	// Get the positional argument for the Fluentd server address
	args := pflag.Args()
	if len(args) < 1 {
		log.Fatal("The Fluentd server address is required as a positional argument (e.g. tcp://127.0.0.1:24224, udp://127.0.0.1:54453, or unix:///run/fluentd.sock).")
	}
	address := args[0]

	// If --record is not provided, generate a default record dynamically
	if *recordInput == "" {
		userName := getCurrentUser()
		hostName := getHostname()

		record := fmt.Sprintf(`{
			"message": "This is a test event",
			"local": {
				"user": "%s",
				"host": "%s",
				"address": "%s"
			}
		}`, userName, hostName, address)

		recordInput = &record
	}

	// Parse the record input
	record, err := parseRecord(*recordInput)
	if err != nil {
		log.Fatalf("Failed to parse record input: %v", err)
	}

	// Parse the address and create a Fluent Forward client
	network, addr, err := parseAddress(address)
	if err != nil {
		log.Println("Address format should match any of tcp:(/..)<host>:<port>, udp:(/..)<host>:<port>, or unix:(/..)<path>")
		log.Println("Any number of slashes is optional; use an odd number for absolute unix paths")
		log.Fatalf("Failed to parse address: %v", err)
	}
	fmt.Printf("Connecting to Fluentd at %s://%s\n", network, addr)
	client := client.New(client.ConnectionOptions{
		Factory: &client.ConnFactory{
			Network: network,
			Address: addr,
			Timeout: 5 * time.Second,
		},
	})

	// Ensure the client is connected
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to Fluentd: %v", err)
	}
	defer client.Disconnect()

	// Send the message
	if err := client.SendMessage(*tag, record); err != nil {
		log.Fatalf("Failed to send event: %v", err)
	}

	log.Println("Event sent successfully")
}

// getCurrentUser returns the current username or "unknown" if it cannot be determined
func getCurrentUser() string {
	user, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return user.Username
}

// getHostname returns the hostname or "unknown" if it cannot be determined
func getHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}

// parseAddress parses the address string and returns the network and address
func parseAddress(address string) (string, string, error) {
	// Merged regex pattern for tcp, udp, and unix with distinct matching rules
	pattern := `^(?:(?P<scheme>tcp|udp):(?P<slashes>/+)?(?P<authority>.+)|(?P<scheme>unix):(?P<slashes>//)*(?P<authority>/+.*))$`
	regex := regexp.MustCompile(pattern)

	// Validate match and extract network and address
	if match := regex.FindStringSubmatch(address); match != nil {
		if match[1] != "" { // tcp or udp
			return match[1], match[3], nil
		} else if match[4] != "" { // unix
			return match[4], match[6], nil
		} else {
			return "", "", fmt.Errorf("invalid address format")
		}
	}
	return "", "", fmt.Errorf("invalid format")
}

// parseRecord parses a record, either as a JSON string or from a file prefixed with '@'
func parseRecord(input string) (map[string]interface{}, error) {
	var data interface{}

	if strings.HasPrefix(input, "@") {
		// File path mode
		filePath := strings.TrimPrefix(input, "@")
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
		}
		err = json.Unmarshal(fileContent, &data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON from file %s: %v", filePath, err)
		}
	} else {
		// Treat input as JSON string
		err := json.Unmarshal([]byte(input), &data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON string: %v", err)
		}
	}

	// Ensure the data is a map (object) and not an array
	record, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("record must be a JSON object, not an array or other type")
	}

	return record, nil
}
