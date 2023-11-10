# Mattermost Weather Integration

## Overview

This Go application serves as a backend for a Mattermost slash command that provides real-time weather information. When a user types `/weather [location]` in Mattermost, this application fetches weather data for the specified location and returns it to the Mattermost channel.

## Features

- Fetches real-time weather information using an external API.
- Integrates seamlessly with Mattermost as a custom slash command.
- Easy configuration via an external JSON file.

## Prerequisites

Before you begin, ensure you have the following:

- Go (version 1.20 or higher)
- Mattermost server access to register the slash command
- An API key from a weather data provider (e.g., OpenWeatherMap, WeatherAPI)

## Installation

1. Clone the Repository:
```bash
git clone https://github.com/jlandells/mm-weather
cd mm-weather
```

2. Configuration:
    - Rename config.json.example to config.json.
    - Edit config.json to include your weather API key:
```json

{
    "apiKey": "your_api_key_here"
}
```

3. Building the Application:
```go
go build
```

4. Running the Application:
```bash
./mm-weather_[platform-specific-suffix]
```

## Usage

1. Register the Slash Command in Mattermost:
    - In your Mattermost instance, go to the integrations section and add a new slash command.
    - Set the command trigger word, e.g., weather.
    - Point the request URL to where this Go application is hosted with the /weather endpoint.
2. Using the Command in Mattermost:
    - In any Mattermost channel, type `/weather [location]` to get the weather information for the specified location.

## Contributing

Contributions to this project are welcome! Please fork the repository and submit a pull request with your changes or improvements.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details, or visit [https://www.gnu.org/licenses/gpl-3.0.html](https://www.gnu.org/licenses/gpl-3.0.html).

## Contact

For questions, feedback, or contributions regarding this project, please use the following methods:

- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements, feel free to open an issue or a pull request in this repository.
- **Mattermost Community**: Join us in the Mattermost Community server, where we discuss all things related to extending Mattermost. You can find us in the channel [Integrations and Apps](https://community.mattermost.com/core/channels/integrations).
- Social Media: Follow and message me on Twitter, where I'm [@jlandells](https://twitter.com/jlandells).