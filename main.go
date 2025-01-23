package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mattermost/mattermost/server/public/model"
)

var Version = "development" // Default value - overwritten during bild process

var debugMode bool = false

// LogLevel is used to refer to the type of message that will be written using the logging code.
type LogLevel string

type mmConnection struct {
	mmURL    string
	mmPort   string
	mmScheme string
	mmToken  string
}

const (
	debugLevel   LogLevel = "DEBUG"
	infoLevel    LogLevel = "INFO"
	warningLevel LogLevel = "WARNING"
	errorLevel   LogLevel = "ERROR"
)

const (
	defaultPort       = "443"
	defaultScheme     = "https"
	pageSize          = 60
	maxErrors         = 3
	maxMessageLength  = 40
	menuPostPerPage   = 2
	conf_file_default = "config.json"
)

type PostSummary struct {
	PostID  string
	Message string
}

// Logging functions

// LogMessage logs a formatted message to stdout or stderr
func LogMessage(level LogLevel, message string) {
	if level == errorLevel {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("[%s] %s\n", level, message)
}

// DebugPrint allows us to add debug messages into our code, which are only printed if we're running in debug more.
// Note that the command line parameter '-debug' can be used to enable this at runtime.
func DebugPrint(message string) {
	if debugMode {
		LogMessage(debugLevel, message)
	}
}

// getEnvWithDefaults allows us to retrieve Environment variables, and to return either the current value or a supplied default
func getEnvWithDefault(key string, defaultValue interface{}) interface{} {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func main() {

	// Parse Command Line
	DebugPrint("Parsing command line")

	var MattermostURL string
	var MattermostPort string
	var MattermostScheme string
	var MattermostToken string
	var MattermostChannel string
	var ConfigFilename string
	var NoHeaderFlag bool
	var DebugFlag bool
	var VersionFlag bool

	flag.StringVar(&MattermostURL, "url", "", "The URL of the Mattermost instance (without the HTTP scheme)")
	flag.StringVar(&MattermostPort, "port", "", "The TCP port used by Mattermost. [Default: "+defaultPort+"]")
	flag.StringVar(&MattermostScheme, "scheme", "", "The HTTP scheme to be used (http/https). [Default: "+defaultScheme+"]")
	flag.StringVar(&MattermostToken, "token", "", "The auth token used to connect to Mattermost")
	flag.StringVar(&MattermostChannel, "channel", "", "The channel ID to target. (Available from 'Channel Info' screen)")
	flag.StringVar(&ConfigFilename, "config", conf_file_default, "Alternative JSON filename. [Default: "+conf_file_default+"]")
	flag.BoolVar(&NoHeaderFlag, "noheader", false, "Don't create a channel header - just add bookmarks")
	flag.BoolVar(&DebugFlag, "debug", debugMode, "Enable debug output")
	flag.BoolVar(&VersionFlag, "version", false, "Show version information and exit")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Utility to quickly add predefined structures to customer channels in Mattermost.")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	debugMode = DebugFlag

	if VersionFlag {
		fmt.Printf("mm-channel-header - Version: %s\n\n", Version)
		os.Exit(0)
	}

	// If information not supplied on the command line, check whether it's available as an envrionment variable
	if MattermostURL == "" {
		MattermostURL = getEnvWithDefault("MM_URL", "").(string)
	}
	if MattermostPort == "" {
		MattermostPort = getEnvWithDefault("MM_PORT", defaultPort).(string)
	}
	if MattermostScheme == "" {
		MattermostScheme = getEnvWithDefault("MM_SCHEME", defaultScheme).(string)
	}
	if MattermostToken == "" {
		MattermostToken = getEnvWithDefault("MM_TOKEN", "").(string)
	}
	if !DebugFlag {
		DebugFlag = getEnvWithDefault("MM_DEBUG", debugMode).(bool)
	}

	DebugMessage := fmt.Sprintf("Parameters: \n  MattermostURL=%s\n  MattermostPort=%s\n  MattermostScheme=%s\n  MattermostToken=%s\n  ChannelID=%s\n  JSON File=%s\n",
		MattermostURL,
		MattermostPort,
		MattermostScheme,
		MattermostToken,
		MattermostChannel,
		ConfigFilename,
	)
	DebugPrint(DebugMessage)

	// Validate required parameters
	DebugPrint("Validating parameters")
	var cliErrors bool = false
	if MattermostURL == "" {
		LogMessage(errorLevel, "The Mattermost URL must be supplied either on the command line of vie the MM_URL environment variable")
		cliErrors = true
	}
	if MattermostScheme == "" {
		LogMessage(errorLevel, "The Mattermost HTTP scheme must be supplied either on the command line of vie the MM_SCHEME environment variable")
		cliErrors = true
	}
	if MattermostToken == "" {
		LogMessage(errorLevel, "The Mattermost auth token must be supplied either on the command line of vie the MM_TOKEN environment variable")
		cliErrors = true
	}
	if MattermostChannel == "" {
		LogMessage(errorLevel, "A Mattermost Channel ID is required to use this utility.")
		cliErrors = true
	}

	if cliErrors {
		flag.Usage()
		os.Exit(1)
	}

	// Prepare the Mattermost connection
	mattermostConenction := mmConnection{
		mmURL:    MattermostURL,
		mmPort:   MattermostPort,
		mmScheme: MattermostScheme,
		mmToken:  MattermostToken,
	}

	mmTarget := fmt.Sprintf("%s://%s:%s", mattermostConenction.mmScheme, mattermostConenction.mmURL, mattermostConenction.mmPort)

	DebugPrint("Full target for Mattermost: " + mmTarget)
	mmClient := model.NewAPIv4Client(mmTarget)
	mmClient.SetToken(mattermostConenction.mmToken)
	DebugPrint("Connected to Mattermost")

	LogMessage(infoLevel, "Processing started - Version: "+Version)

	config := ProcessConfigFile(ConfigFilename)

	linkToPinnedPost := ProcessPinnedPosts(*mmClient, MattermostChannel, config)

	DebugPrint("Link to pinned post: " + linkToPinnedPost)

	if len(linkToPinnedPost) > 0 {
		if len(config.Bookmarks) > 0 {
			config.Bookmarks = append(config.Bookmarks, Bookmark{
				DisplayName: "Additional Resources",
				LinkURL:     linkToPinnedPost,
				Emoji:       ":bulb:",
			})
		}
	}

	// Only process the channel header if we need to
	if !NoHeaderFlag {
		ProcessChannelHeader(*mmClient, MattermostChannel, config)
	}

	ProcessChannelBookmarks(*mmClient, MattermostChannel, config)

}
