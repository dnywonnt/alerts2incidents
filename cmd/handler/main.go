package main // dnywonnt.me/alerts2incidents/cmd/handler

func main() {
	// Initialize the handler by calling InitializeHandler function which sets up
	// logging, configurations, database connections, caching mechanisms,
	// data channels, and prepares collectors for data gathering.
	handler := InitializeHandler()

	// Start the main processing of the handler which includes:
	// - Starting goroutines for updating caches.
	// - Collecting data through defined collectors.
	// - Processing alerts.
	// - Handling graceful shutdown signals.
	handler.Run()
}
