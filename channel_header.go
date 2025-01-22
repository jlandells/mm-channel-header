package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

func ChannelHeaderExists(mmClient model.Client4, channelID string) (bool, error) {
	DebugPrint("Checking for existing channel header")

	ctx := context.Background()
	etag := ""

	channel, response, err := mmClient.GetChannel(ctx, channelID, etag)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve channel header: "+err.Error())
		return false, err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Function call to GetChannel returned bad HTTP response")
		return false, errors.New("bad HTTP response")
	}

	if len(channel.Header) > 0 {
		return true, nil
	}

	return false, nil
}

func CreateChannelHeader(mmClient model.Client4, channelID string, config *Config) error {

	DebugPrint("Creating channel header")

	ctx := context.Background()

	channelHeader := "Important Data (hover for expanded view)\n\n"

	for _, person := range config.Team {
		personRow := fmt.Sprintf("%s - [%s](%s)\n\n", person.Role, person.Name, person.Email)
		channelHeader += personRow
	}

	channelHeader += "| Key Resources |\n"
	channelHeader += "| -- |\n"

	for _, bookmark := range config.Bookmarks {
		bookmarkRow := fmt.Sprintf("|[%s](%s)|\n", bookmark.DisplayName, bookmark.LinkURL)
		channelHeader += bookmarkRow
	}

	channelPayload := &model.ChannelPatch{
		Header: &channelHeader,
	}

	_, response, err := mmClient.PatchChannel(ctx, channelID, channelPayload)

	if err != nil {
		LogMessage(errorLevel, "Failed to update channel header: "+err.Error())
		return err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Function call to PatchChannel returned bad HTTP response")
		return errors.New("bad HTTP response")
	}

	return nil
}

func ProcessChannelHeader(mmClient model.Client4, MattermostChannel string, config *Config) {
	DebugPrint("Processing channel header")

	if len(config.Bookmarks) == 0 {
		LogMessage(warningLevel, "No bookmarks found in JSON file")
		return
	}
	numBookmarks := fmt.Sprintf("Found %d bookmarks", len(config.Bookmarks))
	DebugPrint(numBookmarks)

	hasHeader, err := ChannelHeaderExists(mmClient, MattermostChannel)
	if err != nil {
		LogMessage(errorLevel, "Unable to validate if channel header exists!  Aborting.")
		os.Exit(8)
	}

	input := "y"

	if hasHeader {
		fmt.Printf("A channel header already exists.  Overwrite? (Press Y to confirm, or any other key to abort)")

		reader := bufio.NewReader(os.Stdin)
		input, err = reader.ReadString('\n')
		if err != nil {
			LogMessage(errorLevel, "Error reading input.  Aborting.")
			os.Exit(9)
		}
	}

	input = strings.TrimSpace(input)
	if strings.ToLower(input) == "y" {
		LogMessage(infoLevel, "Replacing existing Channel Header")
		err = CreateChannelHeader(mmClient, MattermostChannel, config)
		if err != nil {
			LogMessage(errorLevel, "Error creating channel header.  Aborting")
			os.Exit(31)
		}
	} else {
		LogMessage(infoLevel, "Using existing Channel Header")
	}

}
