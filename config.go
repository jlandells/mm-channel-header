package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Struct definitions
type Person struct {
	Role  string `json:"role"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Bookmark struct {
	DisplayName string `json:"display_name"`
	LinkURL     string `json:"link_url"`
	Emoji       string `json:"emoji"`
}

type Resource struct {
	DisplayName string `json:"display_name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Config struct {
	Team      []Person   `json:"team"`
	Bookmarks []Bookmark `json:"bookmarks"`
	Resources []Resource `json:"resources"`
}

// LoadConfig reads the JSON file
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return &config, nil
}

// ProcessTeam processes the team section if it exists
func ProcessTeam(team []Person) {
	if len(team) == 0 {
		LogMessage(infoLevel, "No team information provided in JSON file.")
		return
	}
	if debugMode {
		fmt.Println("  Team Members:")
		for _, person := range team {
			fmt.Printf("    - %s (%s): %s\n", person.Role, person.Name, person.Email)
		}
	}
}

// ProcessBookmarks processes bookmarks if they exist
func ProcessBookmarks(bookmarks []Bookmark) {
	if len(bookmarks) == 0 {
		LogMessage(infoLevel, "No bookmarks provided in JSON file.")
		return
	}
	if debugMode {
		fmt.Println("  Bookmarks:")
		for _, bookmark := range bookmarks {
			fmt.Printf("    - [%s](%s) %s\n", bookmark.DisplayName, bookmark.LinkURL, bookmark.Emoji)
		}
	}
}

// ProcessResources processes resources if they exist
func ProcessResources(resources []Resource) {
	if len(resources) == 0 {
		LogMessage(infoLevel, "No resources provided in JSON file.")
		return
	}
	if debugMode {
		fmt.Println("  Resources:")
		for _, resource := range resources {
			fmt.Printf("    - %s: %s (%s)\n", resource.DisplayName, resource.Description, resource.URL)
		}
	}
}

func ProcessConfigFile(ConfigFilename string) *Config {
	DebugPrint("Processing JSON")

	// Load JSON data
	config, err := LoadConfig(ConfigFilename)
	if err != nil {
		errMesg := fmt.Sprintf("Error processing JSON file: %v", err)
		LogMessage(errorLevel, errMesg)
		os.Exit(12)
	}

	// Process each JSON section
	ProcessTeam(config.Team)
	ProcessBookmarks(config.Bookmarks)
	ProcessResources(config.Resources)

	DebugPrint("JSON processed")

	return config
}
