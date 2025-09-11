package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"repo/internal/say"
	"slices"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var ConfigFile string
var Values config

const (
	StyleFlat      string = "flat"
	StyleRecursive string = "recursive"
)

func Init(flags *pflag.FlagSet) {
	initFailsafecheck()

	var k = koanf.New(".")

	initLoadDefaults(k)
	initLoadConfigfile(k)
	initLoadEnvs(k)
	initLoadFlags(k, flags)

	k.Unmarshal("", &Values)
	validate()

	// Pretty print for debugging
	// s, _ := json.MarshalIndent(i, "", "\t")
	// say.InfoLn("%s", s)
}

// failsafe check for old env-vars
func initFailsafecheck() {
	_, exists := os.LookupEnv("REPOW_STYLE")
	if exists {
		say.Error("REPOW_STYLE has been deprecated, please use REPOW_OPTIONS_STYLE")
		os.Exit(1)
	}
}

func initLoadDefaults(k *koanf.Koanf) {
	k.Load(structs.Provider(config{
		Options: options{
			Style:            "flat",
			Parallelism:      32,
			OptionalManifest: false,
			OptionalContacts: false,
		},
		Server: server{
			Port: 8080,
		},
		Gitlab: gitlab{
			DownloadRetryCount: 6, // lower values didn't solve the issue
			Host:               "gitlab.com",
		},
	}, "koanf"), nil)
	print(k, "loaded defaults")
}

func initLoadConfigfile(k *koanf.Koanf) {
	selectedConfigFile := selectedConfigFile()
	if _, err := os.Stat(selectedConfigFile); err == nil {
		if err := k.Load(file.Provider(selectedConfigFile), yaml.Parser()); err != nil {
			say.Error("error loading config: %v", err)
		}
		print(k, "loaded config-file")
	} else {
		say.Verbose("Config-File %s does not exist\n", selectedConfigFile)
	}
}

func initLoadEnvs(k *koanf.Koanf) {
	k.Load(env.Provider("REPOW_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "REPOW_")), "_", ".", -1)
	}), nil)
	print(k, "loaded envs")
}

// Load flags (and do some ugly mapping)
func initLoadFlags(k *koanf.Koanf, flags *pflag.FlagSet) {
	p := posflag.ProviderWithValue(flags, ".", k, func(key string, value string) (string, any) {
		mappings := map[string]string{
			"optionalContacts": "options.optionalcontacts",
			"optionalManifest": "options.optionalmanifest",
			"parallelism":      "options.parallelism",
			"style":            "options.style",
		}
		if len(mappings[key]) > 0 {
			return mappings[key], value
		}
		return key, value
	})
	// Load flags with provider
	if err := k.Load(p, nil); err != nil {
		log.Fatalf("error loading flags: %v", err)
	}
	print(k, "loaded flags")
}

func DefaultConfigFile() string {
	// points in most cases to "${HOME}/.confg/repow/repow.yaml"
	configDir, err := os.UserConfigDir()
	cobra.CheckErr(err)
	return filepath.Join(filepath.Join(configDir, "repow"), "repow.yaml")
}

func selectedConfigFile() string {
	if len(ConfigFile) > 0 {
		return ConfigFile
	}
	return DefaultConfigFile()
}

func validate() error {
	stylesAvailable := []string{StyleFlat, StyleRecursive}
	if !slices.Contains(stylesAvailable, Values.Options.Style) {
		return fmt.Errorf("invalid value for style: %q", Values.Options.Style)
	}
	return nil
}

func print(k *koanf.Koanf, stage string) {
	say.Verbose(">>> %s", stage)
	var kk = k.Copy()
	kk.Set("gitlab.apitoken", sensitive(k.String("gitlab.apitoken")))
	kk.Set("gitlab.secrettoken", sensitive(k.String("gitlab.secrettoken")))
	say.Verbose("%s", kk.Sprint())
}

func sensitive(value string) string {
	if len(value) > 0 {
		return "(set)"
	}
	return "(not set)"
}
