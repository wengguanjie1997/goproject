package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	AWS AwsConfig `yaml:"aws"`
	TG  TgConfig  `yaml:"tg"`
	HF  HfConfig  `yaml:"hf"`
}
type AwsConfig struct {
	AwsAccessKeyId     string `yaml:"awsAccessKeyId"`
	AwsSecretAccessKey string `yaml:"awsSecretAccessKey"`
}
type TgConfig struct {
	Token string `yaml:"token"`
}
type HfConfig struct {
	ApiKey string `yaml:"apiKey"`
}

func GetConfig() (conf Config, err error) {
	config := &Config{}
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println("Error open the file: ", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return *config, err
	}
	return *config, nil
}
