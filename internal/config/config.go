package config

import (
	"os"
	"path"
	"repo/internal/say"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CfgFile string
var UsedConfig config

// initConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".repow" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".repow")
		CfgFile = path.Join(home, ".repow.yaml")
	}

	if !fileExists(CfgFile) {
		// If the config file does not exist, create it with default values.
		say.InfoLn("Config file %s does not exist, creating with default values", CfgFile)
		if err := writeDefaultConfigFile(CfgFile); err != nil {
			say.Error("Failed to write default config file: %v", err)
			os.Exit(1)
		}
	}

	var replacer = strings.NewReplacer("-", "_", ".", "_")

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(replacer)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		say.InfoLn("Using config file: %s", viper.ConfigFileUsed())
	} else {
		say.Error("Error reading config file: %v", err)
		os.Exit(1)
	}
	viper.Unmarshal(&UsedConfig)
	UsedConfig.log()
}

type config struct {
	Repow struct {
		Server struct {
			Port string
		}
		OptionalContacts bool
	}
	Gitlab struct {
		Host        string
		Token       string
		SecretToken string
	}
	Options struct {
		DownloadRetryCount int
		Style              string
	}
	Slack struct {
		Token     string
		ChannelId string
		Prefix    string
	}
}

func (c *config) log() {
	say.InfoLn("Repow config:")
	say.InfoLn("  Server Port: %s", c.Repow.Server.Port)
	say.InfoLn("  Optional Contacts: %v", c.Repow.OptionalContacts)
	say.InfoLn("  Gitlab Host: %s", c.Gitlab.Host)
	say.InfoLn("  Gitlab Token: %s", "xxx")
	say.InfoLn("  Gitlab Secret Token: %s", "xxx")
	say.InfoLn("  Download Retry Count: %d", c.Options.DownloadRetryCount)
	say.InfoLn("  Style: %s", c.Options.Style)
	say.InfoLn("  Slack Token: %s", "xxx")
	say.InfoLn("  Slack Channel ID: %s", c.Slack.ChannelId)
	say.InfoLn("  Slack Prefix: %s", c.Slack.Prefix)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func defaultConfig() config {
	return config{
		Repow: struct {
			Server struct {
				Port string
			}
			OptionalContacts bool
		}{
			Server: struct{ Port string }{
				Port: "8080",
			},
			OptionalContacts: false,
		},
		Gitlab: struct {
			Host        string
			Token       string
			SecretToken string
		}{
			Host:        "gitlab.com",
			Token:       "",
			SecretToken: "",
		},
		Options: struct {
			DownloadRetryCount int
			Style              string
		}{
			DownloadRetryCount: 6,
			Style:              "flat",
		},
		Slack: struct {
			Token     string
			ChannelId string
			Prefix    string
		}{
			Token:     "",
			ChannelId: "",
			Prefix:    ":large_blue_circle:",
		},
	}
}

func writeDefaultConfigFile(path string) error {
	cfg := defaultConfig()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
