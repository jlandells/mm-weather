package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/buger/jsonparser"
	"github.com/spf13/viper"
)

// Defaults & Type Definitions

var debugMode bool = false
var defaultPort string = "8080"
var weatherAPIKey string

// MattermostResponse represents the key fields that we need to deliver in the
// response to the Mattermost slash command.
type MattermostResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// LogLevel is used to refer to the type of message that will be written using the logging code.
type LogLevel string

const defaultConfigFile string = "config.json"

const (
	debugLevel   LogLevel = "DEBUG"
	infoLevel    LogLevel = "INFO"
	warningLevel LogLevel = "WARNING"
	errorLevel   LogLevel = "ERROR"
)

const (
	weatherAPIBase   string = "http://api.weatherapi.com/v1/current.json"
	weatherAPISuffix string = "aqi=no"
)

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

// Utility functions

// FileExists is used to validate that a file really exists and is an actual file, rather than a directory.
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		LogMessage(errorLevel, filename+" is a directory!")
		return false, nil
	}
	return true, nil
}

// Integration functions

// weatherHandler is the primary function for processing the incoming slash command
func weatherHandler(responseWriter http.ResponseWriter, inboundRequest *http.Request) {
	LogMessage(infoLevel, "Received inbound request")

	// Retrieve the location from the GET request
	location := inboundRequest.URL.Query().Get("text")
	DebugPrint("Text: " + location)
	if location == "" {
		location = "auto:ip"
	}

	// Call the backend API
	apiResponse, err := callWeatherAPI(location)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	currentLocation, _ := jsonparser.GetString([]byte(apiResponse), "location", "name")
	currentTemp, _ := jsonparser.GetFloat([]byte(apiResponse), "current", "temp_c")
	currentConditions, _ := jsonparser.GetString([]byte(apiResponse), "current", "condition", "text")

	// responseMessage contains a Markdown message that gets posted to the channel
	responseMessage := fmt.Sprintf("Current weather in %s: %v°C - %s", currentLocation, currentTemp, currentConditions)

	// responsePayload.ResponseType can be "in_channel" to be posted to the whole channel, or "ephemeral"
	// to be only visible to the person running the slash command.
	responsePayload := MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseMessage,
	}

	// Marshal the response payload to JSON
	jsonResponse, err := json.Marshal(responsePayload)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the JSON response back to Mattermost
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	_, writeErr := responseWriter.Write(jsonResponse)
	if writeErr != nil {
		LogMessage(errorLevel, "Error posting response to Mattermost: "+string(writeErr.Error()))
	}
}

// callWeatherAPI is the code that, as the name suggests, actually calls out to the weather API
func callWeatherAPI(location string) (string, error) {
	DebugPrint("Calling weather API for location: " + location)

	safeLocation := url.QueryEscape(location)

	fullURL := fmt.Sprintf("%s?key=%s&q=%s&%s", weatherAPIBase, weatherAPIKey, safeLocation, weatherAPISuffix)

	// Make the GET call to retrieve the data
	resp, err := http.Get(fullURL)
	if err != nil {
		LogMessage(errorLevel, "Call to Weather API failed")
		return "", err
	}
	defer resp.Body.Close()

	// Extract the body of the message
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogMessage(errorLevel, "Unable to extract body data from Weather API response")
		return "", err
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		LogMessage(errorLevel, "Failed to convert body data")
		return "", err
	}

	// Convert the data to a string to return to the calling function
	weatherData, err := json.Marshal(result)
	if err != nil {
		LogMessage(errorLevel, "Unable to convert weather data to string")
		return "", err
	}

	DebugPrint("Weather data: " + string(weatherData))

	return string(weatherData), nil
}

func main() {
	var debugFlag bool
	var configFile string
	var apiToken string
	var listenPort string

	flag.BoolVar(&debugFlag, "debug", false, "Enable debug mode")
	flag.StringVar(&configFile, "config", "config.json", "Override default config file (config.json)")
	flag.StringVar(&apiToken, "token", "", "Override the API token supplied in the config file")
	flag.StringVar(&listenPort, "port", "", "Override the port that this utility should listen on")

	flag.Parse()

	debugMode = debugFlag
	var exists bool

	// If the API token is not passed on the command line, we should check whether it exists as an
	// environment variable before reading the value from the config file.
	if apiToken == "" {
		// Not set via command line - check environment
		apiToken, exists = os.LookupEnv("WEATHER_API_TOKEN")
		if !exists {
			// Still no API token - let's check the config file
			viper.SetConfigFile(configFile)
			err := viper.ReadInConfig()
			if err != nil {
				panic(fmt.Errorf("fatal error processing config file: %w", err))
			}
			apiToken = viper.GetString("apiKey")
			DebugPrint("Obtained API key from config file")
		} else {
			DebugPrint("Obtained API key from environment")
		}
	} else {
		DebugPrint("Obtained API key from command line")
	}

	if apiToken == "" {
		LogMessage(errorLevel, "Failed to locate API key!")
		os.Exit(2)
	}

	weatherAPIKey = apiToken

	// In the same way that we validated the API key, we need a valid port parameter, except in this case
	// we have a programmatic default
	if listenPort == "" {
		// Not set via command line - check environment
		listenPort, exists = os.LookupEnv("MM_LISTEN_PORT")
		if !exists {
			// Still no API token - let's check the config file (which we should already have!)
			listenPort = viper.GetString("listenPort")
			if listenPort == "" {
				listenPort = defaultPort
				DebugPrint("Using default listen port: " + listenPort)
			} else {
				DebugPrint("Obtained listen port '" + listenPort + "' from config file")
			}
		} else {
			DebugPrint("Obtained listen port '" + listenPort + "' from environment")
		}
	} else {
		DebugPrint("Obtained listen port '" + listenPort + "' from command line")
	}

	// Setup the inbound request handler
	LogMessage(infoLevel, "Starting server on port "+listenPort)
	listenPortString := fmt.Sprintf(":%s", listenPort)
	DebugPrint("Listen port string: " + listenPortString)
	http.HandleFunc("/weather", weatherHandler)
	if err := http.ListenAndServe(listenPortString, nil); err != nil {
		panic(err)
	}

}
