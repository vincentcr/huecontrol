package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/vincentcr/huecontrol/hue"
)

func main() {
	hostname, username := loadConfig()
	client := hue.New(hostname, username) //"10.0.0.15", "38d7fa8a6e94c1718bb02f62203e733"
	runMethod(client, "GetGroups")
	runMethod(client, "GetGroup", "1")
}

func loadConfig() (string, string) {
	contents, err := ioutil.ReadFile("./.config.json")
	if err != nil {
		panic(fmt.Errorf("Unable to read config file: %v", err))
	}

	var config struct {
		Hostname string
		Username string
	}

	err = json.Unmarshal(contents, &config)
	if err != nil {
		panic(fmt.Errorf("Unable to parse config file: %v", err))
	}

	return config.Hostname, config.Username
}

func runMethod(client *hue.Client, name string, args ...interface{}) {
	argsValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argsValues[i] = reflect.ValueOf(arg)
	}
	results := reflect.ValueOf(client).MethodByName(name).Call(argsValues)
	out := results[0].Interface()
	err := results[1].Interface()
	if err != nil {
		panic(err)
	} else {
		outputFormatted(out)
	}
}

func outputFormatted(obj interface{}) error {
	formatted, err := format(obj)
	if err != nil {
		return err
	}
	fmt.Println(formatted)
	return nil
}

func format(obj interface{}) (string, error) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("unable to format for output: %#v: %v", obj, err)
	}

	return string(bytes), nil
}
