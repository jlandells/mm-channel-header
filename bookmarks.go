package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattermost/mattermost/server/public/model"
)

type BookmarkAction string

const (
	BookmarkReplace BookmarkAction = "Replace"
	BookmarkAppend  BookmarkAction = "Append"
	BookmarkAbort   BookmarkAction = "Abort"
)

type BookmarkActionModel struct {
	options  []BookmarkAction
	cursor   int
	selected BookmarkAction
	done     bool
}

func NewBookmarkActionModel() *BookmarkActionModel {
	return &BookmarkActionModel{
		options: []BookmarkAction{
			BookmarkReplace,
			BookmarkAppend,
			BookmarkAbort,
		},
		cursor:   0,
		selected: BookmarkAbort, // Default to "Abort" if no choice is made
		done:     false,
	}
}

func (m *BookmarkActionModel) Init() tea.Cmd {
	return nil
}

func (m *BookmarkActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.done = true
			m.selected = BookmarkAbort
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.done = true
			m.selected = m.options[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *BookmarkActionModel) View() string {
	if m.done {
		return ""
	}

	s := "Existing Bookmarks Found\n\n"
	for i, option := range m.options {
		cursor := " " // No cursor by default
		if i == m.cursor {
			cursor = ">" // Highlight the current option
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}
	s += "\nUse ↑/↓ to navigate and Enter to select. Press q to abort."
	return s
}

func (m *BookmarkActionModel) SelectedAction() BookmarkAction {
	return m.selected
}

func RunBookmarkMenu() BookmarkAction {
	model := NewBookmarkActionModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running bookmark menu: %v\n", err)
		return BookmarkAbort
	}
	return model.SelectedAction()
}

func HasExistingBookmarks(mmClient model.Client4, channelID string) (bool, error) {
	DebugPrint("Checking for existing bookmarks")

	ctx := context.Background()

	bookmarks, response, err := mmClient.ListChannelBookmarksForChannel(ctx, channelID, 0)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve bookmarks: "+err.Error())
		return false, err
	}
	if response.StatusCode != 200 && response.StatusCode != 201 {
		LogMessage(errorLevel, "Function call to ListChannelBookmarksForChannel returned badf HTTP response")
		return false, errors.New("bad HTTP response")
	}

	if len(bookmarks) == 0 {
		return false, nil
	}

	return true, nil
}

func DeleteExistingBookmarks(mmClient model.Client4, channelID string) error {
	DebugPrint("Deleting existing bookmarks")

	ctx := context.Background()

	bookmarks, response, err := mmClient.ListChannelBookmarksForChannel(ctx, channelID, 0)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve bookmarks: "+err.Error())
		return err
	}
	if response.StatusCode != 200 && response.StatusCode != 201 {
		LogMessage(errorLevel, "Function call to ListChannelBookmarksForChannel returned badf HTTP response")
		return errors.New("bad HTTP response")
	}

	for _, bookmark := range bookmarks {
		_, response, err = mmClient.DeleteChannelBookmark(ctx, channelID, bookmark.Id)

		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete bookmark with ID: %s (Name: %s). Error: %s", bookmark.Id, bookmark.DisplayName, err.Error())
			LogMessage(errorLevel, errorMsg)
			return err
		}
		if response.StatusCode != 200 && response.StatusCode != 201 {
			LogMessage(errorLevel, "Function call to DeleteChannelBookmark returned bad HTTP response")
		}

	}

	return nil
}

func CreateBookmarks(mmClient model.Client4, channelID string, config *Config) error {
	DebugPrint("Creating bookmarks")

	ctx := context.Background()

	for _, bookmark := range config.Bookmarks {
		bookmarkPayload := &model.ChannelBookmark{
			ChannelId:   channelID,
			DisplayName: bookmark.DisplayName,
			LinkUrl:     bookmark.LinkURL,
			Emoji:       bookmark.Emoji,
			Type:        "link",
		}

		_, response, err := mmClient.CreateChannelBookmark(ctx, bookmarkPayload)

		if err != nil {
			LogMessage(errorLevel, "Failed to create bookmark: "+err.Error())
			return err
		}
		if response.StatusCode != 200 && response.StatusCode != 201 {
			LogMessage(errorLevel, "Function call to CreateChannelBookmark returned bad HTTP response")
			return errors.New("bad HTTP response")
		}
	}

	return nil
}

func ProcessChannelBookmarks(mmClient model.Client4, channelID string, config *Config) {
	DebugPrint("Processing channel bookmarks")

	if len(config.Bookmarks) == 0 {
		LogMessage(infoLevel, "No bookmarks found in JSON file")
		return
	}

	hasBookmarks, err := HasExistingBookmarks(mmClient, channelID)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve existing bookmarks.  Aborting.")
		os.Exit(41)
	}

	if hasBookmarks {
		action := RunBookmarkMenu()

		switch action {
		case BookmarkReplace:
			LogMessage(infoLevel, "Replacing existing bookbarks")
			err = DeleteExistingBookmarks(mmClient, channelID)
			if err != nil {
				LogMessage(errorLevel, "Failed to delete existing bookmarks.  Aborting.")
				os.Exit(45)
			}
		case BookmarkAppend:
			LogMessage(infoLevel, "Appending bookmarks to existing")
		case BookmarkAbort:
			LogMessage(warningLevel, "Aborting.  Please review existing bookmarks!")
			return
		}
	}
	err = CreateBookmarks(mmClient, channelID, config)

	if err != nil {
		LogMessage(errorLevel, "Failed to create bookmarks.  Aborting.")
		os.Exit(42)
	}

}
