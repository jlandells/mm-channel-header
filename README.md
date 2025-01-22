# Bookmark Utility for Mattermost

This utility is designed to simplify the management of channel bookmarks and headers for a Mattermost instance. It allows you to configure channel bookmarks via a JSON file, interactively handle existing bookmarks, and optionally update the channel header.

---

## Features

- **Add Channel Bookmarks**: Create bookmarks in a Mattermost channel based on a structured JSON file.
- **Interactive Bookmark Handling**: If existing bookmarks are found, choose to replace, append, or abort.
- **Channel Header Update**: Automatically update the channel header with relevant information unless the `-noheader` flag is provided.
- **Configurable via JSON**: Define bookmarks, team details, and additional resources in a single configuration file.

---

## Download

You can download the latest release of this utility from the [GitHub Releases](https://github.com/jlandells/mm-channel-header/releases) page.

1. Navigate to the releases section.
2. Download the appropriate binary for your operating system.
3. Place the binary in your desired location and ensure it has executable permissions.

---

## JSON Configuration

The utility relies on a JSON file to define the details of the bookmarks, team, and resources. Below is an example structure:

```json
{
  "team": {
    "tam": {
      "name": "John Doe",
      "email": "tam@example.com"
    },
    "csm": {
      "name": "Jane Smith",
      "email": "csm@example.com"
    }
  },
  "bookmarks": [
    {
      "display_name": "Documentation",
      "link_url": "https://docs.mattermost.com",
      "emoji": ":book:"
    },
    {
      "display_name": "API Reference",
      "link_url": "https://api.mattermost.com",
      "emoji": ":computer:"
    }
  ],
  "resources": [
    {
      "display_name": "Academy",
      "url": "https://academy.mattermost.com/",
      "description": "Courses to enhance your Mattermost knowledge."
    }
  ]
}
```

---

## Command-Line Parameters

The utility can be configured using command-line options or environment variables. Below is a list of supported parameters:

| **Option**     | **Env Var Alternative** | **Required?** | **Description**                                | **Default**     |
|----------------|--------------------------|---------------|------------------------------------------------|-----------------|
| `-url`         | `MM_URL`                | Yes           | Mattermost instance URL                       |                 |
| `-port`        | `MM_PORT`               | No            | Mattermost port                               | 443             |
| `-scheme`      | `MM_SCHEME`             | No            | The HTTP scheme to be used (`http`/`https`).  | `https`         |
| `-token`       | `MM_TOKEN`              | Yes           | The API token for Mattermost                 |                 |
| `-channel`     |                          | Yes           | Mattermost channel ID                         |                 |
| `-config`      |                          | No            | JSON file containing the config definition    | `config.json`   |
| `-noheader`    |                          | No            | If present, no channel header is created.     |                 |
| `-debug`       | `MM_DEBUG`              | No            | Run the utility in DEBUG mode                | False           |

---

## Examples

### Add Bookmarks and Update the Channel Header
```sh
./mm-channel-header_<os_version> -url https://mattermost.example.com -token YOUR_API_TOKEN -channel CHANNEL_ID -config config.json
```

### Add Bookmarks Without Updating the Channel Header
```sh
./mm-channel-header_<os_version> -url https://mattermost.example.com -token YOUR_API_TOKEN -channel CHANNEL_ID -config config.json -noheader
```

### Enable Debug Mode
```sh
./mm-channel-header_<os_version> -url https://mattermost.example.com -token YOUR_API_TOKEN -channel CHANNEL_ID -debug
```

In all examples, command-line parameters will override corresponding environment variables.

---

## Contributing

We welcome contributions from the community! Whether it's a bug report, a feature suggestion, or a pull request, your input is valuable to us. Please feel free to contribute in the following ways:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements, open an issue or a pull request in this repository.
- **Mattermost Community**: Join the discussion in the [Integrations and Apps](https://community.mattermost.com/core/channels/integrations) channel on the Mattermost Community server.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For questions, feedback, or contributions regarding this project, please use the following methods:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements, feel free to open an issue or a pull request in this repository.
- **Mattermost Community**: Join us in the Mattermost Community server, where we discuss all things related to extending Mattermost. You can find me in the channel [Integrations and Apps](https://community.mattermost.com/core/channels/integrations).
- **Social Media**: Follow and message me on Twitter, where I'm [@jlandells](https://twitter.com/jlandells).


