package websocket_notify

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

type Config struct {
	ApiListen struct {
		Host    string
		Port    uint16
		SSLCert string
		SSLKey  string
		SSL     bool
	}

	WsListen struct {
		Host    string
		Port    uint16
		SSLCert string
		SSLKey  string
		SSL     bool
	}

	ApiSecret string
	WsSecret  string
}

type YamlConfig struct {
	ApiListen struct {
		Host    string `yaml:"host"`
		Port    uint16 `yaml:"port"`
		SSLCert string `yaml:"ssl_cert"`
		SSLKey  string `yaml:"ssl_key"`
	} `yaml:"api"`

	WsListen struct {
		Host    string `yaml:"host"`
		Port    uint16 `yaml:"port"`
		SSLCert string `yaml:"ssl_cert"`
		SSLKey  string `yaml:"ssl_key"`
	} `yaml:"websocket"`

	ApiSecret string `yaml:"api_secret"`
	WsSecret  string `yaml:"websocket_secret"`
}

func loadConfig(cmd *cobra.Command) Config {
	yamlConfig := YamlConfig{}
	yamlConfig.ApiListen.Host = `0.0.0.0`
	yamlConfig.ApiListen.Port = 8000
	yamlConfig.WsListen.Host = `0.0.0.0`
	yamlConfig.WsListen.Port = 8001

	configFilePath := cmd.Flag(`config`).Value.String()
	configFileYaml, err := os.ReadFile(configFilePath)
	if err == nil {
		err = yaml.Unmarshal(configFileYaml, &yamlConfig)
		if err != nil {
			log.Fatalln(`Failed to parse config file. Error: `, err)
		}
	}

	config := Config{}

	config.ApiSecret = getValue(cmd, `api-secret`, `API_SECRET`, yamlConfig.ApiSecret)
	config.WsSecret = getValue(cmd, `ws-secret`, `WS_SECRET`, yamlConfig.ApiSecret)

	config.ApiListen.Host = getValue(cmd, `api-host`, `API_HOST`, yamlConfig.ApiListen.Host)
	config.ApiListen.Port = getValueInt(cmd, `api-port`, `API_PORT`, yamlConfig.ApiListen.Port)
	config.ApiListen.SSLCert = getValue(cmd, `api-ssl-cert`, `API_SSL_CERT`, yamlConfig.ApiListen.SSLCert)
	config.ApiListen.SSLKey = getValue(cmd, `api-ssl-key`, `API_SSL_KEY`, yamlConfig.ApiListen.SSLKey)

	config.WsListen.Host = getValue(cmd, `ws-host`, `API_HOST`, yamlConfig.WsListen.Host)
	config.WsListen.Port = getValueInt(cmd, `ws-port`, `API_PORT`, yamlConfig.WsListen.Port)
	config.WsListen.SSLCert = getValue(cmd, `ws-ssl-cert`, `API_SSL_CERT`, yamlConfig.WsListen.SSLCert)
	config.WsListen.SSLKey = getValue(cmd, `ws-ssl-key`, `API_SSL_KEY`, yamlConfig.WsListen.SSLKey)

	config.CheckSSL()

	return config
}

func getValue(cmd *cobra.Command, cmdFlag string, envName string, yamlValue string) string {
	value := yamlValue

	envValue, envOk := os.LookupEnv(fmt.Sprintf(`WEBSOCKET_NOTIFY_%s`, envName))
	if envOk {
		value = envValue
	}

	if cmd.Flags().Changed(cmdFlag) {
		value, _ = cmd.Flags().GetString(cmdFlag)
	}

	return value
}

func getValueInt(cmd *cobra.Command, cmdFlag string, envName string, yamlValue uint16) uint16 {
	value := yamlValue

	envValue, envOk := os.LookupEnv(fmt.Sprintf(`WEBSOCKET_NOTIFY_%s`, envName))
	if envOk {
		intValue, err := strconv.ParseUint(envValue, 10, 16)

		if err != nil {
			log.Fatalln(fmt.Sprintf(`Environment variable %s is not a number.`, envName))
		}

		value = uint16(intValue)
	}

	if cmd.Flags().Changed(cmdFlag) {
		value, _ = cmd.Flags().GetUint16(cmdFlag)
	}

	return value
}

func (config *Config) CheckSSL() {
	if config.ApiListen.SSLCert != `` && config.ApiListen.SSLKey == `` {
		log.Fatalln(`API SSL certificate is defined, but the SSL key is not.`)
	}

	if config.ApiListen.SSLCert == `` && config.ApiListen.SSLKey != `` {
		log.Fatalln(`API SSL key is defined, but the SSL certificate is not.`)
	}

	config.ApiListen.SSL = config.ApiListen.SSLCert != ``

	if config.WsListen.SSLCert != `` && config.WsListen.SSLKey == `` {
		log.Fatalln(`WebSocket SSL certificate is defined, but the SSL key is not.`)
	}

	if config.WsListen.SSLCert == `` && config.WsListen.SSLKey != `` {
		log.Fatalln(`WebSocket SSL key is defined, but the SSL certificate is not.`)
	}

	config.WsListen.SSL = config.WsListen.SSLCert != ``
}
