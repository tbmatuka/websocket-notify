websocket-notify
===============

Pass events from a HTTP API to authenticated websocket listeners.

# Features

* Hash signed auth that doesn't hit your backend
* HTTP API for sending tagged or broadcast events
* Websocket for subscribing to event tags

# Configuration

## Flags
```
  --api-host string       API host to listen on, can be an interface name (default "0.0.0.0")
  --api-port uint16       API port to listen on (default 8000)
  --api-secret string     secret string used to authorize API requests
  --api-ssl-cert string   API SSL certificate path
  --api-ssl-key string    API SSL certificate key path
  --config string         config file path (default "/etc/websocket-notify.yaml")
  -h, --help              help for websocket-notify
  --ws-host string        WebSocket host to listen on, can be an interface name (default "0.0.0.0")
  --ws-port uint16        WebSocket port to listen on (default 8001)
  --ws-secret string      secret string used to authorize WebSocket subscriptions
  --ws-ssl-cert string    WebSocket SSL certificate path
  --ws-ssl-key string     WebSocket SSL certificate key path
```

## YAML config file
```yaml
api_secret: ''
websocket_secret: ''

api:
    host: 0.0.0.0
    port: 8000
    ssl_cert: /path/to/ssl.crt
    ssl_key: /path/to/ssl.key

websocket:
    host: 0.0.0.0
    port: 8001
    ssl_cert: /path/to/ssl.crt
    ssl_key: /path/to/ssl.key

debug: true
```

## Environment variables

```shell
WEBSOCKET_NOTIFY_API_SECRET=""
WEBSOCKET_NOTIFY_WS_SECRET=""

WEBSOCKET_NOTIFY_API_HOST="0.0.0.0"
WEBSOCKET_NOTIFY_API_PORT="8000"
WEBSOCKET_NOTIFY_API_SSL_CERT=""
WEBSOCKET_NOTIFY_API_SSL_KEY=""

WEBSOCKET_NOTIFY_WS_HOST="0.0.0.0"
WEBSOCKET_NOTIFY_WS_PORT="8001"
WEBSOCKET_NOTIFY_WS_SSL_CERT=""
WEBSOCKET_NOTIFY_WS_SSL_KEY=""

WEBSOCKET_NOTIFY_DEBUG="true"
```
