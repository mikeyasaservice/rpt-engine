package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"scout/engine"
)

func main() {
	// Open a log file
	f, err := os.OpenFile("scout.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()

	// Set log output to the file
	log.SetOutput(f)

	// Define the configuration for your ruleset
	config := engine.Config{
		Directory:       []string{"/Users/mikesaxton/Development/scout/rules"},
		FailOnRuleParse: true,
		FailOnYamlParse: true,
		NoCollapseWS:    false,
	}

	log.Println("Initializing ruleset...")

	// Initialize a new ruleset
	ruleset, err := engine.NewRuleset(config, []string{})
	if err != nil {
		log.Fatalf("Failed to initialize ruleset: %v", err)
	}

	log.Printf("Loaded %d rules\n", ruleset.Ok)

	if ruleset.Ok == 0 {
		log.Fatalf("No rules loaded. Please check the rules directory path and file format.")
	}

	// Print the loaded rules for verification
	for _, rule := range ruleset.Rules {
		fmt.Printf("Loaded Rule ID: %s, Title: %s\n", rule.Rule.ID, rule.Rule.Title)
	}

	// Read logs from a file
	logsFile := "logs.json"
	logsData, err := ioutil.ReadFile(logsFile)
	if err != nil {
		log.Fatalf("Failed to read logs file: %v", err)
	}

	var events []engine.SampleEvent
	if err := json.Unmarshal(logsData, &events); err != nil {
		log.Fatalf("Failed to unmarshal logs data: %v", err)
	}

	for _, event := range events {
		log.Printf("Evaluating event: %+v\n", event)
		results, matched := ruleset.EvalAll(event)
		if matched {
			fmt.Println("Event matched the following rules:")
			for _, result := range results {
				log.Printf("Rule ID: %s\nTitle: %s\nDescription: %s\n", result.ID, result.Title, result.Description)
				fmt.Printf("Rule ID: %s\nTitle: %s\nDescription: %s\n\n", result.ID, result.Title, result.Description)
			}
		} else {
			log.Println("No rules matched the event.")
			fmt.Println("No rules matched the event.")
		}
	}

	// Handle errors and alerts
	log.Println("Handling errors and alerts...")
	if ruleset.Failed > 0 {
		log.Println("Errors occurred during rule processing:")
		// Assuming you have a way to get the errors from the ruleset
	} else {
		log.Println("No errors occurred during rule processing.")
	}
	log.Println("Finished handling errors and alerts.")
}
