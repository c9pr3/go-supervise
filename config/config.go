package config

import (
	"encoding/json"
	"fmt"
	"log/syslog"
	"os"
)

var LOGGER, LOGERR = syslog.New(syslog.LOG_WARNING, "svscan")

type Configuration struct {
	ServiceConfig struct {
		Path                    string `json:"path"`
		MaxFailedStartups       int    `json:"max_failed_startups"`
		TimeWaitBetweenStartups int    `json:"time_between_startups"`
	}
	LogConfig struct {
		LogLevel  string `json:"level"` // not used yet
		LogEmpty  bool   `json:"log_empty"`
		LogSyslog bool   `json:"log_syslog"`
		Tai64     bool   `json:"tai64"`
	}
	LogFilesConfig struct {
		MaxFiles    int    `json:"max"`
		DeleteFiles int    `json:"delete"`
		MaxFileSize uint32 `json:"size"`
	}
}

func ReadConfig() (Configuration, error) {
	file, _ := os.Open("./../config/config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		LOGGER.Crit(fmt.Sprintf("Could not parse config. Error: %s", err))
		return configuration, err
	}

	return configuration, nil
}
