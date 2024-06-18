package database // dnywonnt.me/alerts2incidents/internal/database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

// ListenChannel defines a type for channel names used for listening to notifications.
type ListenChannel string

// Define constants for channel names to listen to for notifications.
const (
	IncidentsChannel ListenChannel = "a2i_incidents_channel" // Channel for incident notifications.
	RulesChannel     ListenChannel = "a2i_rules_channel"     // Channel for rule notifications.
)

// ListenToNotifications listens for notifications on a PostgreSQL channel and triggers a handler function when a notification is received.
func ListenToNotifications(ctx context.Context, pool *pgxpool.Pool, channel ListenChannel, handleFunc func(notification *pgconn.Notification)) {
	// Log the start of the notification listener with the channel it is listening to
	log.WithFields(log.Fields{
		"channel": channel,
	}).Debug("Starting the notification listener")

	// Continuously try to listen for notifications
	for {
		// Attempt to acquire a connection from the pool
		conn, err := pool.Acquire(ctx)
		if err != nil {
			// Log the failure to acquire a connection and retry after a delay
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to acquire a connection from the pool; will retry")
			time.Sleep(time.Second * 10)
			continue
		}

		// Execute the LISTEN command to listen on the specified channel
		if _, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s", channel)); err != nil {
			// Log the failure to initiate listening and release the connection
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"channel": channel,
			}).Error("Failed to listen to the channel; will retry")
			conn.Release()
			continue
		}

		// Notify that the listener has successfully started
		log.WithFields(log.Fields{
			"channel": channel,
		}).Debug("Notification listener has successfully started")

		// Loop to wait for and handle incoming notifications
		for {
			// Wait for a notification on the connection
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// Check if the context was canceled, indicating we should stop listening
				if errors.Is(err, context.Canceled) {
					log.WithFields(log.Fields{
						"channel": channel,
					}).Debug("Stopping the notification listener")
					conn.Release()
					return
				}

				// Log any errors while waiting for notifications and break out to try reconnecting
				log.WithFields(log.Fields{
					"error":   err.Error(),
					"channel": channel,
				}).Error("Failed to wait for a notification; will attempt to reconnect")
				conn.Release()
				break
			}

			// Handle the received notification using the provided handler function
			log.WithFields(log.Fields{
				"channel": channel,
				"payload": notification.Payload,
			}).Debug("New notification received; starting handler function")
			handleFunc(notification)
		}
	}
}
