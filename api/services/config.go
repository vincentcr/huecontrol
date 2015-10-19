package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/imdario/mergo"
	"github.com/kardianos/osext"
)

const (
	defaultConfigFilename = "config.json"
	defaultConfigKey      = "default"
)

type Config struct {
	PublicURL   string
	PostgresURL string
	RedisURL    string
}

func loadConfig(env string) (Config, error) {
	configFile, err := defaultConfigFile()
	if err != nil {
		return Config{}, err
	}
	return loadConfigFromFile(env, configFile)
}

func defaultConfigFile() (string, error) {
	dirs, err := configDirs()
	if err != nil {
		return "", err
	}

	for _, dir := range dirs {
		fullPath := path.Join(dir, defaultConfigFilename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	cwd, _ := os.Getwd()
	return "", fmt.Errorf("could not find config file in dirs %v (cwd: %v)", dirs, cwd)
}

func configDirs() ([]string, error) {
	exeDir, err := exeDir()
	if err != nil {
		return nil, err
	}

	dirs := []string{".", exeDir, ".."}
	return dirs, nil
}

func exeDir() (string, error) {
	exe, err := osext.Executable()
	if err != nil {
		return "", err
	}
	return path.Dir(exe), nil
}

func loadConfigFromFile(env string, path string) (Config, error) {
	log.Printf("Loading config for env:%v at path:%v", env, path)
	var configs map[string]Config

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to read file at path '%v': %v", path, err)
	}
	err = json.Unmarshal(contents, &configs)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to read file at path '%v': %v", path, err)
	}

	config, err := mergeConfig(env, configs)
	if err != nil {
		return Config{}, err
	}

	return config, err
}

func mergeConfig(env string, configs map[string]Config) (Config, error) {
	config := configs[env]
	defaultConfig := configs["default"]
	if err := mergo.Merge(&config, defaultConfig); err != nil {
		return config, fmt.Errorf("Failed to merge config for env '%v', %#v: %v", env, configs, err)
	}

	return config, nil
}
