package websocket_notify

import (
	"os"
)

func Execute() {
	rootCmd := getServeCmd()

	rootCmd.PersistentFlags().String(`config`, `/etc/websocket-notify.yaml`, `config file path`)

	rootCmd.PersistentFlags().String(`api-secret`, ``, `secret string used to authorize API requests`)

	rootCmd.PersistentFlags().String(`api-host`, `0.0.0.0`, `API host to listen on, can be an interface name`)
	rootCmd.PersistentFlags().Uint16(`api-port`, 8000, `API port to listen on`)
	rootCmd.PersistentFlags().String(`api-ssl-cert`, ``, `API SSL certificate path`)
	rootCmd.PersistentFlags().String(`api-ssl-key`, ``, `API SSL certificate key path`)

	rootCmd.PersistentFlags().String(`ws-secret`, ``, `secret string used to authorize WebSocket subscriptions`)

	rootCmd.PersistentFlags().String(`ws-host`, `0.0.0.0`, `WebSocket host to listen on, can be an interface name`)
	rootCmd.PersistentFlags().Uint16(`ws-port`, 8001, `WebSocket port to listen on`)
	rootCmd.PersistentFlags().String(`ws-ssl-cert`, ``, `WebSocket SSL certificate path`)
	rootCmd.PersistentFlags().String(`ws-ssl-key`, ``, `WebSocket SSL certificate key path`)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
