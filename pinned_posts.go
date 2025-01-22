package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattermost/mattermost/server/public/model"
)

// PaginationModel handles the interactive selection menu with pagination
type PaginationModel struct {
	posts       []PostSummary
	currentPage int
	perPage     int
	selection   int
	done        bool
	selected    string
}

type SelectionResult struct {
	SelectionType string // "PinnedPost", "AddNew", "Skip", or "Abort"
	PostID        string // The ID of the selected post (if applicable)
	Message       string // The message of the selected post (if applicable)
}

// Functions for handling interactive menus
func NewPaginationModel(posts []PostSummary, perPage int) *PaginationModel {
	return &PaginationModel{
		posts:   posts,
		perPage: perPage,
	}
}

func (m *PaginationModel) Init() tea.Cmd {
	return nil
}

func (m *PaginationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.done = true
			m.selected = "Abort"
			m.selection = -1
			return m, tea.Quit
		case "enter":
			selected := m.visibleOptions()[m.selection]
			switch selected {
			case "Next page":
				if (m.currentPage+1)*m.perPage < len(m.posts) {
					m.currentPage++
					m.selection = 0
				}
			case "Previous page":
				if m.currentPage > 0 {
					m.currentPage--
					m.selection = 0
				}
			case "Add a new pinned post", "Abort", "Skip":
				m.done = true
				m.selected = selected
				return m, tea.Quit
			default:
				// Post selected
				m.done = true
				m.selected = selected
				return m, tea.Quit
			}
		case "up":
			if m.selection > 0 {
				m.selection--
			}
		case "down":
			if m.selection < len(m.visibleOptions())-1 {
				m.selection++
			}
		case "left":
			if m.currentPage > 0 {
				m.currentPage--
				m.selection = 0
			}
		case "right":
			if (m.currentPage+1)*m.perPage < len(m.posts) {
				m.currentPage++
				m.selection = 0
			}
		}
	}

	return m, nil
}

func (m *PaginationModel) View() string {
	if m.done {
		return ""
	}

	options := m.visibleOptions()
	s := "Use one of these existing pinned posts, or add a new one?\n\n"
	for i, option := range options {
		cursor := " " // No cursor
		if i == m.selection {
			cursor = ">" // Highlight the selected option
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}

	s += "\nUse ↑/↓ to navigate, ←/→ to switch pages, and Enter to select. Press q to abort.\n"
	return s
}

func (m *PaginationModel) visibleOptions() []string {
	// Calcualte the range of posts for the current page
	start := m.currentPage * m.perPage
	end := start + m.perPage
	if end > len(m.posts) {
		end = len(m.posts)
	}

	// Populate the menu for the current page
	options := []string{}
	for _, post := range m.posts[start:end] {
		options = append(options, post.Message)
	}

	// Add navigation and special options
	if m.currentPage > 0 {
		options = append(options, "Previous page")
	}
	if end < len(m.posts) {
		options = append(options, "Next page")
	}
	options = append(options, "Add a new pinned post")
	options = append(options, "Abort")
	options = append(options, "Skip")
	return options
}

func (m *PaginationModel) SelectedOption() string {
	options := m.visibleOptions()
	if m.selection >= 0 && m.selection < len(options) {
		return options[m.selection]
	}
	return "Abort"
}

func BuildLinkToPinnedPost(mmClient model.Client4, channelID string, postID string) (string, error) {
	DebugPrint("Building link to pinned post")

	linkToPinnedPost := ""

	baseURL := mmClient.URL

	ctx := context.Background()
	etag := ""

	// We need to get the Team name to build the URL.  We can get the Team ID from the Channel,
	// then lookup the Team name from there.

	channel, response, err := mmClient.GetChannel(ctx, channelID, etag)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve channel data: "+err.Error())
		return "", err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Function call to GetChannel returned badf HTTP response")
		return "", errors.New("bad HTTP response")
	}

	DebugPrint("Found Team ID: " + channel.TeamId)

	team, response, err := mmClient.GetTeam(ctx, channel.TeamId, etag)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve team data: "+err.Error())
		return "", err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Function call to GetTeam returned badf HTTP response")
		return "", errors.New("bad HTTP response")
	}

	linkToPinnedPost = fmt.Sprintf("%s/%s/pl/%s", baseURL, team.Name, postID)

	return linkToPinnedPost, nil
}

func GetPinnedPost(mmClient model.Client4, channelID string) (SelectionResult, error) {
	DebugPrint("Retrieving pinned posts")

	ctx := context.Background()
	etag := ""

	pinned_posts, response, err := mmClient.GetPinnedPosts(ctx, channelID, etag)

	if err != nil {
		LogMessage(errorLevel, "Failed to retrieve pinned posts: "+err.Error())
		return SelectionResult{}, err
	}
	if response.StatusCode != 200 {
		LogMessage(errorLevel, "Function call to GetPinnedPosts returned bad HTTP response")
		return SelectionResult{}, errors.New("bad HTTP response")
	}

	if len(pinned_posts.Order) <= 0 {
		DebugPrint("No pinned posts found")
		return SelectionResult{SelectionType: "AddNew"}, nil
	}

	var postSummaries []PostSummary
	for _, postID := range pinned_posts.Order {
		post := pinned_posts.Posts[postID]
		splitMessage := strings.Split(post.Message, "\n")[0]
		shortMessage := truncateString(splitMessage, maxMessageLength)
		postSummaries = append(postSummaries, PostSummary{
			PostID:  post.Id,
			Message: shortMessage,
		})
	}

	menuModel := NewPaginationModel(postSummaries, menuPostPerPage)

	// Run the interactive menu
	p := tea.NewProgram(menuModel)
	if _, err := p.Run(); err != nil {
		LogMessage(errorLevel, "Error displaying menu: "+err.Error())
		return SelectionResult{}, err
	}

	selected := menuModel.SelectedOption()

	for _, post := range postSummaries {
		truncatedMessage := truncateString(post.Message, maxMessageLength)
		if selected == truncatedMessage {
			return SelectionResult{
				SelectionType: "PinnedPost",
				PostID:        post.PostID,
				Message:       post.Message,
			}, nil
		}
	}

	// Handle Special Cases
	switch selected {
	case "Add a new pinned post":
		return SelectionResult{SelectionType: "AddNew"}, nil
	case "Abort":
		return SelectionResult{SelectionType: "Abort"}, nil
	case "Skip":
		return SelectionResult{SelectionType: "Skip"}, nil
	}

	// Fallback (this should never be reached!)
	return SelectionResult{SelectionType: "Error"}, errors.New("error processing interactive menu")
}

func truncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:(maxLength-3)] + "..."
	}
	return input
}

func CreatePinnedPost(mmClient model.Client4, channelID string, jsonRows Config) (string, error) {
	DebugPrint("Creating pinned post from JSON data")

	ctx := context.Background()

	pinnedPostMessage := "## Additional Resources\n\n\n"
	pinnedPostMessage += "| Resource                                                                                                        | Description                                                                                                     |\n"
	pinnedPostMessage += "| --------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |\n"

	for _, row := range jsonRows.Resources {
		formattedRow := fmt.Sprintf("| [%s](%s) | %s |\n", row.DisplayName, row.URL, row.Description)
		pinnedPostMessage += formattedRow
	}

	postPayload := &model.Post{
		ChannelId: channelID,
		IsPinned:  true,
		Message:   pinnedPostMessage,
	}

	post, response, err := mmClient.CreatePost(ctx, postPayload)

	if err != nil {
		LogMessage(errorLevel, "Failed to create pinned post: "+err.Error())
		return "", err
	}
	// Note that we're looking for an HTTP 201 response for this, rather than the more usual 200
	if response.StatusCode != 201 {
		LogMessage(errorLevel, "Function call to CreatePost returned bad HTTP response")
		return "", errors.New("bad HTTP response")
	}

	return post.Id, nil
}

func ProcessPinnedPosts(mmClient model.Client4, MattermostChannel string, config *Config) string {

	pinnedPost, err := GetPinnedPost(mmClient, MattermostChannel)

	if err != nil {
		LogMessage(errorLevel, "Interactive menu failed - aborting!")
		os.Exit(4)
	}

	pinnedPostID := pinnedPost.PostID

	switch pinnedPost.SelectionType {
	case "PinnedPost":
		DebugPrint("Existing Pinned Post selected.  Post ID: " + pinnedPostID)
	case "AddNew":
		LogMessage(infoLevel, "Adding new post from JSON")
		pinnedPostID, err = CreatePinnedPost(mmClient, MattermostChannel, *config)
		if err != nil {
			LogMessage(errorLevel, "Failed to create pinned post!  "+err.Error())
			os.Exit(6)
		}
	case "Skip":
		LogMessage(infoLevel, "Skipping pinned post")
		return ""
	case "Abort":
		LogMessage(warningLevel, "Aborting due to user selection")
		os.Exit(0)
	default:
		LogMessage(errorLevel, "Interactive menu got funky!  This code should never be reached!! ( ˶°ㅁ°) !!")
		os.Exit(3)
	}

	DebugPrint("Pinned Post ID: " + pinnedPostID)

	linkToPinnedPost, err := BuildLinkToPinnedPost(mmClient, MattermostChannel, pinnedPostID)

	if err != nil {
		LogMessage(errorLevel, "Failed to build link to pinned post - Aborting.")
		os.Exit(7)
	}

	return linkToPinnedPost
}
