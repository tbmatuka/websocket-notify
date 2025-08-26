package websocket_notify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

type SubscriptionManager struct {
	connections      map[uint]*SocketConnection
	tags             map[string]map[uint]*SocketConnection
	connectionsMutex sync.Mutex
	tagsMutex        sync.Mutex
	secret           []byte
}

var (
	subscriptionManagerLock = &sync.Mutex{}      //nolint:gochecknoglobals
	subscriptionManager     *SubscriptionManager //nolint:gochecknoglobals
)

func getSubscriptionManager() *SubscriptionManager {
	if subscriptionManager == nil {
		// only lock on initialization
		subscriptionManagerLock.Lock()

		// check again after lock
		if subscriptionManager == nil {
			subscriptionManager = new(SubscriptionManager)
			subscriptionManager.connections = make(map[uint]*SocketConnection)
			subscriptionManager.tags = make(map[string]map[uint]*SocketConnection)
		}

		subscriptionManagerLock.Unlock()
	}

	return subscriptionManager
}

var errInvalidSignature = errors.New(`invalid signature`)

func (manager *SubscriptionManager) Subscribe(connection *SocketConnection, tags []string, signature string) error {
	if len(manager.secret) != 0 {
		correctSignature := hex.EncodeToString(
			pbkdf2.Key(
				[]byte(strings.Join(tags, `|`)),
				manager.secret,
				10,
				16,
				sha256.New,
			),
		)

		if signature != correctSignature {
			return errInvalidSignature
		}
	}

	if !connection.Managed {
		manager.addConnection(connection)
	}

	manager.subscribeToTags(connection, tags)

	return nil
}

func (manager *SubscriptionManager) Unsubscribe(connection *SocketConnection, tags []string) {
	manager.unsubscribeFromTags(connection, tags)
}

func (manager *SubscriptionManager) CloseConnection(connection *SocketConnection) {
	if len(connection.Tags) > 0 {
		tags := make([]string, 0, len(connection.Tags))

		for tagName := range connection.Tags {
			tags = append(tags, tagName)
		}

		manager.unsubscribeFromTags(connection, tags)
	}

	manager.removeConnection(connection)
}

func (manager *SubscriptionManager) DistributeEvent(event Event) {
	eventString, err := json.Marshal(event)
	if err != nil {
		return
	}

	for _, tagName := range event.Tags {
		for _, connection := range manager.tags[tagName] {
			connection.Channel <- eventString
		}
	}
}

func (manager *SubscriptionManager) Status() ApiStatus {
	subscriptions := make(map[string]uint)

	for tag, connections := range manager.tags {
		subscriptions[tag] = uint(len(connections))
	}

	return ApiStatus{
		Connections:   uint(len(manager.connections)),
		Subscriptions: subscriptions,
	}
}

func (manager *SubscriptionManager) subscribeToTags(connection *SocketConnection, tags []string) {
	manager.tagsMutex.Lock()
	defer manager.tagsMutex.Unlock()

	for _, tagName := range tags {
		if connection.Tags[tagName] {
			continue
		}

		_, tagExists := manager.tags[tagName]

		if !tagExists {
			manager.tags[tagName] = make(map[uint]*SocketConnection)
		}

		logger.Debug(fmt.Sprintf(`Adding connection %d to tag %s`, connection.Id, tagName))
		manager.tags[tagName][uint(len(manager.tags[tagName]))] = connection
		connection.Tags[tagName] = true
	}
}

func (manager *SubscriptionManager) unsubscribeFromTags(connection *SocketConnection, tags []string) {
	manager.tagsMutex.Lock()
	defer manager.tagsMutex.Unlock()

	for _, tagName := range tags {
		if len(manager.tags[tagName]) == 1 {
			delete(manager.tags, tagName)
		} else {
			for index, tagConnection := range manager.tags[tagName] {
				if tagConnection == connection {
					logger.Debug(fmt.Sprintf(`Removing connection %d from tag %s`, connection.Id, tagName))

					manager.tags[tagName][index] = manager.tags[tagName][uint(len(manager.tags[tagName]))-1]
					delete(manager.tags[tagName], uint(len(manager.tags[tagName]))-1)

					break
				}
			}
		}

		delete(connection.Tags, tagName)
	}
}

func (manager *SubscriptionManager) addConnection(connection *SocketConnection) {
	manager.connectionsMutex.Lock()
	defer manager.connectionsMutex.Unlock()

	connection.Id = uint(len(manager.connections))
	connection.Managed = true
	manager.connections[connection.Id] = connection

	logger.Debug(fmt.Sprintf(`Added connection %d`, connection.Id))
}

func (manager *SubscriptionManager) removeConnection(connection *SocketConnection) {
	manager.connectionsMutex.Lock()
	defer manager.connectionsMutex.Unlock()

	lastConnectionId := uint(len(manager.connections)) - 1

	logger.Debug(fmt.Sprintf(`Removing connection %d`, connection.Id))

	if connection.Id != lastConnectionId {
		logger.Debug(fmt.Sprintf(`Swapping connection %d with %d`, connection.Id, lastConnectionId))
		lastConnection := manager.connections[lastConnectionId]
		lastConnection.Id = connection.Id

		manager.connections[connection.Id] = lastConnection
	}

	delete(manager.connections, lastConnectionId)
	logger.Debug(fmt.Sprintf(`Removed connection %d`, lastConnectionId))

	connection.Id = uint(0)
	connection.Managed = false
}
