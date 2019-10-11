package main

import (
	"Assignment1Summary/servers/gateway/handlers"
	"log"
	"net/http"
	"os"
)

// The main function is the main entry point for the server
// It essentially handles the paths each of the go files need to work for
func main() {
	/* TODO: add code to do the following
	- Read the ADDR environment variable to get the address
	  the server should listen on. If empty, default to ":80"
	- Create a new mux for the web server.
	- Tell the mux to call your handlers.SummaryHandler function
	  when the "/v1/summary" URL path is requested.
	- Start a web server listening on the address you read from
	  the environment variable, using the mux you created as
	  the root handler. Use log.Fatal() to report any errors
	  that occur when trying to start the web server.
	*/

	//Getting the value of the ADDR environment variable
	addr := os.Getenv("ADDR")

	//If ADDR is blank, default to ":80"
	if len(addr) == 0 {
		addr = ":80"
	}

	//Create a mux
	mux := http.NewServeMux()
	//Assign the SummaryHandler function to the mux path "/v1/summary"
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	//Start the web zipserver and report errors
	//We use Log.Fatal as this is command line utliity otherwise 
	//we would use http.Error() to respond to a client
	log.Printf("server is listening at https://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))

}
