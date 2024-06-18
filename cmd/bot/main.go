package main //dnywonnt.me/alerts2incidents/cmd/bot

func main() {
	// Initialize the bot using the InitializeBot function
	// This function configures and returns a new instance of Bot with all dependencies set up
	bot := InitializeBot()

	// Start the bot's operation
	// The Run method includes the main logic to handle notifications to Telegram chats
	bot.Run()
}
