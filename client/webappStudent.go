// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// This code initializes and runs the WebApp on the Pi client

package main

import (
	"flag"
	"fmt"
	"PiScanStudent/client/database"
	"PiScanStudent/client/ui"
	"log"
	"net/http"
	"path"
)

const (
	// server constants (runs on Pi client)
	SERVER_HOST = ""
	SERVER_PORT = 8080

	// API server constants (remote server)
	API_HOST = "https://api.saruzai.com"
	API_PORT = 443

	// non-html mime types (ajax replies)
	MIME_JSON = "application/json"
)

func main() {
	var (
		host, apiHost, templatesFolder, dbPath, dbFile string
		port, apiPort                                  int
	)
	flag.StringVar(&host, "host", SERVER_HOST, fmt.Sprintf("Host name or IP address for this server (defaults to '%s')", SERVER_HOST))
	flag.IntVar(&port, "port", SERVER_PORT, fmt.Sprintf("Port addess for this server (defaults to '%d')", SERVER_PORT))
	flag.StringVar(&apiHost, "apiHost", API_HOST, "Host name or IP address for the API server (REQUIRED)")
	flag.IntVar(&apiPort, "apiPort", API_PORT, fmt.Sprintf("Port addess for the API server (defaults to '%d')", API_PORT))
	flag.StringVar(&templatesFolder, "templates", "", "Path to the html templates (REQUIRED)")
	flag.StringVar(&dbPath, "dbPath", database.SQLITE_PATH, fmt.Sprintf("Path to the sqlite file (defaults to '%s')", database.SQLITE_PATH))
	flag.StringVar(&dbFile, "dbFile", database.SQLITE_FILE, fmt.Sprintf("The sqlite database file (defaults to '%s')", database.SQLITE_FILE))
	flag.Parse()

	// make sure the required parameters are passed when run
	if templatesFolder == "" || apiHost == "" {
		fmt.Println("WebApp usage:")
		flag.PrintDefaults()
	} else {
		/* set the server ready for use */
		// confirm the html templates
		ui.InitializeTemplates(templatesFolder)

		// coordinates for connecting to the sqlite database (from the command line options)
		dbCoordinates := database.ConnCoordinates{DBPath: dbPath, DBFile: dbFile}

		// prepare the apiHost:apiPort for handler functions that need them
		extraCoordinates := make([]interface{}, 1)
		extraCoordinates[0] = fmt.Sprintf("%s:%d", apiHost, apiPort)

		/* define the server handlers */
		// dynamic request handlers: html
		http.HandleFunc("/", ui.Redirect("/scanned/"))
		http.HandleFunc("/browser", ui.UnsupportedBrowserHandler(templatesFolder))
		http.HandleFunc("/shutdown/", ui.ShutdownClientHandler())
		http.HandleFunc("/submission/", ui.MakeHTMLHandler(ui.CreatedStudents, dbCoordinates))
		http.HandleFunc("/delete/", ui.MakeHTMLHandler(ui.DeleteStudent, dbCoordinates))
		http.HandleFunc("/input/", ui.MakeHTMLHandler(ui.InputUnknownStudent, dbCoordinates, extraCoordinates...))

		// ajax
		http.HandleFunc("/remove/", ui.MakeHandler(ui.RemoveSingleSubmission, dbCoordinates, MIME_JSON))

		// static resources
		http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(templatesFolder, "../css/")))))
		http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(templatesFolder, "../js/")))))
		http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(path.Join(templatesFolder, "../fonts/")))))
		http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(path.Join(templatesFolder, "../images/")))))

		/* start the server */
		log.Println(fmt.Sprintf("Starting the WebApp %s", fmt.Sprintf("%s:%d", host, port)))
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
	}
}
