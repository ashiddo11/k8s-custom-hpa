package config

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Apps []struct {
		Name   string `yaml:"name"`
		Query  string `yaml:"query"`
		ScaleFactor int32 `yaml:"scaleFactor"`
		MaxReplicas int32 `yaml:"maxReplicas"`
		MinReplicas int32 `yaml:"minReplicas"`
		Namespace	string `yaml:"namespace"`
	} `yaml:"apps"`
}

func LoadConfig() (c *config, err error) {
	log.Printf("Reading config from config.yaml")

	configData, err := ioutil.ReadFile("/config/config.yaml")
	if err != nil {
		log.Fatal("unable to read config.yaml, does the file exist?")
	}

	log.Printf("config.yaml read successfully")

	err = yaml.Unmarshal([]byte(configData), &c)
	if err != nil {
		log.Fatal("Unable to parse config.yaml file")
	}
	log.Printf("config.yaml parsed successfully")
	log.Printf("config: %+v", c)
	return
}
