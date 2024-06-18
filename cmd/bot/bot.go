package main // dnywonnt.me/alerts2incidents/cmd/bot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"dnywonnt.me/alerts2incidents/internal/cache"
	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/database"
	"dnywonnt.me/alerts2incidents/internal/database/repositories"
	"dnywonnt.me/alerts2incidents/internal/models"
	"dnywonnt.me/alerts2incidents/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mymmrac/telego"

	log "github.com/sirupsen/logrus"
)

// Bot represents the Telegram bot instance
type Bot struct {
	cfg           *config.TelegramBotConfig
	dbPool        *pgxpool.Pool
	tgo           *telego.Bot
	incidentsRepo *repositories.IncidentsRepository
	messageCache  *cache.Cache
	messageTmpl   *template.Template
}

// InitializeBot initializes and returns a new Bot instance
func InitializeBot() *Bot {
	log.Info("Initializing the bot")

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load database config")
	}

	// Load Telegram bot configuration
	tgConfig, err := config.LoadTelegramBotConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load Telegram bot config")
	}

	// Create connection string for PostgreSQL
	encodedPassword := url.QueryEscape(dbConfig.Password)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d",
		dbConfig.User, encodedPassword, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.MaxConnections)
	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to create database pool")
	}

	// Load message template from file
	tmpl, err := loadMessageTemplateFromFile(tgConfig.MessageTemplateFilepath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"filePath": tgConfig.MessageTemplateFilepath,
		}).Fatal("Failed to load message template file")
	}

	// Create a new Telegram bot instance
	tgo, err := telego.NewBot(tgConfig.Token, telego.WithDiscardLogger())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to create Telegram bot instance")
	}

	// Return a new Bot instance
	return &Bot{
		cfg:           tgConfig,
		dbPool:        dbPool,
		tgo:           tgo,
		incidentsRepo: repositories.NewIncidentsRepository(dbPool),
		messageCache:  cache.NewCache(tgConfig.MessageCacheMaxSize, "messages"),
		messageTmpl:   tmpl,
	}
}

// Run starts the bot and listens for signals to stop it
func (bot *Bot) Run() {
	log.WithFields(log.Fields{
		"requestDelay": bot.cfg.RequestDelay.String(),
	}).Info("Starting the bot")

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start listening to database notifications
	go database.ListenToNotifications(ctx, bot.dbPool, database.IncidentsChannel, func(notification *pgconn.Notification) {
		if err := bot.handleMessagesForNotification(ctx, notification); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to handle messages for the database notification")
		}
	})

	log.WithFields(log.Fields{
		"requestDelay": bot.cfg.RequestDelay.String(),
	}).Info("The bot successfully started; waiting for incidents")

	// Listen for system signals for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	log.Info("Stopping the bot")

	// Cancel the context, close db pool and clear cache
	cancel()
	bot.dbPool.Close()
	bot.messageCache.Clear()

	log.Info("The bot has been stopped")
}

// handleMessagesForNotification processes notifications from the database
func (bot *Bot) handleMessagesForNotification(ctx context.Context, notification *pgconn.Notification) error {
	// Split the notification payload into action and ID
	parts := strings.SplitN(notification.Payload, ":", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid payload in notification: %s", notification.Payload)
	}

	action, id := parts[0], parts[1]

	// Handle different actions based on the notification
	switch action {
	case "INSERT", "UPDATE":
		// Fetch the incident details from the repository
		incident, err := bot.incidentsRepo.GetIncident(ctx, id)
		if err != nil {
			return fmt.Errorf("error getting incident: %w", err)
		}

		// Render the incident to message text
		text, err := bot.renderIncidentToMessageText(incident)
		if err != nil {
			return fmt.Errorf("error rendering incident model to message text: %w", err)
		}

		// Send or update messages based on the action
		if action == "INSERT" {
			bot.sendMessagesForIncident(incident.ID, text)
		} else if action == "UPDATE" {
			bot.updateMessagesForIncident(incident.ID, text)
		}

	case "DELETE":
		// Handle delete messages
		bot.deleteMessagesForIncident(id)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

// sendMessagesForIncident sends messages to all configured chats
func (bot *Bot) sendMessagesForIncident(incidentID, text string) {
	messages := []*telego.Message{}

	for _, chatStr := range bot.cfg.Chats {
		// Parse chat ID and thread ID from configuration
		chatID, threadID, err := parseChatStr(chatStr)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err.Error(),
				"chatStr": chatStr,
			}).Error("Failed to parse chat string")
			continue
		}

		// Prepare parameters for sending the message
		sendParams := &telego.SendMessageParams{
			ChatID:    telego.ChatID{ID: chatID},
			Text:      text,
			ParseMode: bot.cfg.MessageParseMode,
		}
		if threadID != nil {
			sendParams.MessageThreadID = *threadID
		}

		// Send the message
		msg, err := bot.tgo.SendMessage(sendParams)
		if err != nil {
			logFields := log.Fields{
				"error":  err.Error(),
				"chatID": chatID,
			}
			if threadID != nil {
				logFields["threadID"] = *threadID
			}
			log.WithFields(logFields).Error("Failed to send a message to the chat")
			continue
		}

		logFields := log.Fields{
			"chatID":     chatID,
			"incidentID": incidentID,
		}
		if threadID != nil {
			logFields["threadID"] = *threadID
		}
		log.WithFields(logFields).Info("The message has been sent to the chat for incident")

		messages = append(messages, msg)

		// Adding a delay between requests to prevent exceeding rate limits
		time.Sleep(bot.cfg.RequestDelay)
	}

	// Cache the messages sent for the incident
	if len(messages) > 0 {
		bot.messageCache.SetItem(incidentID, messages)
	}
}

// updateMessagesForIncident updates messages for an existing incident
func (bot *Bot) updateMessagesForIncident(incidentID, text string) {
	item, exists := bot.messageCache.GetItem(incidentID)
	if !exists {
		log.WithFields(log.Fields{
			"incidentID": incidentID,
		}).Warn("The messages not found in the cache for incident; skipping update")
		return
	}
	messages, ok := item.Value.([]*telego.Message)
	if !ok {
		return
	}

	// Update each message in the chat
	for _, msg := range messages {
		if _, err := bot.tgo.EditMessageText(&telego.EditMessageTextParams{
			ChatID:    telego.ChatID{ID: msg.Chat.ID},
			MessageID: msg.MessageID,
			Text:      text,
			ParseMode: bot.cfg.MessageParseMode,
		}); err != nil {
			log.WithFields(log.Fields{
				"error":     err.Error(),
				"chatID":    msg.Chat.ID,
				"messageID": msg.MessageID,
			}).Error("Failed to update a message in the chat")
			continue
		}

		log.WithFields(log.Fields{
			"incidentID": incidentID,
			"chatID":     msg.Chat.ID,
			"messageID":  msg.MessageID,
		}).Info("The message has been updated for incident")

		// Adding a delay between requests to prevent exceeding rate limits
		time.Sleep(bot.cfg.RequestDelay)
	}
}

// deleteMessagesForIncident deletes messages for a deleted incident
func (bot *Bot) deleteMessagesForIncident(incidentID string) {
	item, exists := bot.messageCache.GetItem(incidentID)
	if !exists {
		log.WithFields(log.Fields{
			"incidentID": incidentID,
		}).Warn("The messages not found in the cache for incident; skipping deletion")
		return
	}
	defer bot.messageCache.DeleteItem(incidentID)

	messages, ok := item.Value.([]*telego.Message)
	if !ok {
		return
	}

	// Delete each message in the chat
	for _, msg := range messages {
		if err := bot.tgo.DeleteMessage(&telego.DeleteMessageParams{
			ChatID:    telego.ChatID{ID: msg.Chat.ID},
			MessageID: msg.MessageID,
		}); err != nil {
			log.WithFields(log.Fields{
				"error":     err.Error(),
				"chatID":    msg.Chat.ID,
				"messageID": msg.MessageID,
			}).Error("Failed to delete a message from the chat")
			continue
		}

		log.WithFields(log.Fields{
			"incidentID": incidentID,
			"chatID":     msg.Chat.ID,
			"messageID":  msg.MessageID,
		}).Info("The message has been deleted for incident")

		// Adding a delay between requests to prevent exceeding rate limits
		time.Sleep(bot.cfg.RequestDelay)
	}
}

// renderIncidentToMessageText renders an incident to a message text using the template
func (bot *Bot) renderIncidentToMessageText(incident *models.Incident) (string, error) {
	buf := bytes.Buffer{}
	if err := bot.messageTmpl.Execute(&buf, incident); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// loadMessageTemplateFromFile loads the message template from a file
func loadMessageTemplateFromFile(filepath string) (*template.Template, error) {
	tmpl := template.New("messageTemplate").Funcs(template.FuncMap{
		"joinWithCommas":      utils.JoinWithCommas,
		"escapeMDV2":          utils.EscapeMarkdownV2,
		"derefStr":            utils.DerefStr,
		"prettyJSON":          utils.PrettyJSON,
		"formatNumWithCommas": utils.FormatNumberWithCommas,
	})

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %w", err)
	}

	return tmpl, nil
}

// parseChatStr parses a chat string into chat ID and thread ID
func parseChatStr(chatStr string) (int64, *int, error) {
	parts := strings.SplitN(chatStr, ":", 2)
	if len(parts) < 1 {
		return 0, nil, errors.New("invalid chat string format")
	}
	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("error parsing chatID string to int64: %w", err)
	}

	var threadID *int
	if len(parts) == 2 {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, nil, fmt.Errorf("error parsing threadID string to int: %w", err)
		}
		threadID = &id
	}
	return chatID, threadID, nil
}
