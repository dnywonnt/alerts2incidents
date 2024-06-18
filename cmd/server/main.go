package main // dnywonnt.me/alerts2incidents/cmd/server

func main() {
	// Initialize the server using the InitializeServer function, which sets up all necessary configurations,
	// database connections, routes, and middleware.
	server := InitializeServer()

	// Start the server and handle its lifecycle including graceful shutdown through the Run method.
	// This method will keep running until it receives a termination signal like SIGINT or SIGTERM.
	server.Run()
}
