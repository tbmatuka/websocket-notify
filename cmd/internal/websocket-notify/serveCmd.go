package websocket_notify

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func getServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `websocket-notify`,
		Short: `Pass events from a HTTP API to websocket listeners`,
		Run:   runServeCmd,
	}

	return cmd
}

func runServeCmd(cmd *cobra.Command, _ []string) {
	config := loadConfig(cmd)

	getLogger()
	logger.SetDebug(config.Debug)
	logger.Debug(`Debug enabled`)

	manager := getSubscriptionManager()
	manager.secret = []byte(config.WsSecret)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		host := getListenAddress(config.ApiListen.Host)

		server := &http.Server{
			Addr:              fmt.Sprintf(`%s:%d`, host, config.ApiListen.Port),
			ReadHeaderTimeout: 3 * time.Second,
		}

		server.Handler = http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.URL.Path == `/event` {
				handleEventRequest(responseWriter, request, config)
			}

			if request.URL.Path == `/status` {
				handleStatusRequest(responseWriter, request, config)
			}
		})

		log.Println(`API Listening on: `, server.Addr)

		if config.ApiListen.SSL {
			logger.Debug(`API SSL enabled`)
			log.Fatal(server.ListenAndServeTLS(config.ApiListen.SSLCert, config.ApiListen.SSLKey))
		} else { //nolint:revive
			logger.Debug(`API SSL disabled`)
			log.Fatal(server.ListenAndServe())
		}

		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {
		host := getListenAddress(config.WsListen.Host)

		server := &http.Server{
			Addr:              fmt.Sprintf(`%s:%d`, host, config.WsListen.Port),
			ReadHeaderTimeout: 3 * time.Second,
		}

		server.Handler = http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if request.URL.Path == `/ws` {
				handleWebsocketRequest(responseWriter, request)
			}
		})

		log.Println(`Websocket Listening on: `, server.Addr)

		if config.WsListen.SSL {
			logger.Debug(`Websocket SSL enabled`)
			log.Fatal(server.ListenAndServeTLS(config.WsListen.SSLCert, config.WsListen.SSLKey))
		} else { //nolint:revive
			logger.Debug(`Websocket SSL disabled`)
			log.Fatal(server.ListenAndServe())
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()
}

func getListenAddress(host string) string {
	networkInterface, err := net.InterfaceByName(host)
	if err == nil {
		addresses, _ := networkInterface.Addrs()
		firstAddress, ok := addresses[0].(*net.IPNet)

		if !ok {
			log.Fatalln(`Failed to get address for interface: `, networkInterface.Name)
		}

		return firstAddress.IP.String()
	}

	return host
}
