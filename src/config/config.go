package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbDriver             string `json:"db_driver"`
	DbHost               string `json:"db_host"`
	DbPort               string `json:"db_port"`
	DbUser               string `json:"db_user"`
	DbPasswd             string `json:"db_passwd"`
	DbName               string `json:"db_name"`
	BaseDir              string `json:"base_dir"`
	PublicImagesDir      string `json:"public_images_dir"`
	Oauth2FbClientID     string `json:"oauth2_fb_client_id"`
	Oauth2FbClientSecret string `json:"oauth2_fb_client_secret"`
	Oauth2FbRedirectUrl  string `json:"oauth2_fb_redirect_url"`
}

func ReadDbConfig(inputConfigFile string) Config {
	configFile, err := os.Open(inputConfigFile)
	defer configFile.Close()
	if err != nil {
		panic("Error opening config file")
	}

	decoder := json.NewDecoder(configFile)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		panic("Error decoding config file")
	}

	return config
}

func (c *Config) GetConnString() string {

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.DbUser,
		c.DbPasswd,
		c.DbHost,
		c.DbPort,
		c.DbName)

	return connString
}
