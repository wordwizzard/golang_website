package host

import ("fmt"
		"net/http"
		"html/template"
		"github.com/gorilla/mux"
		"github.com/wordwizzard/playin_go_web/tag"
		"github.com/wordwizzard/playin_go_web/sse"
)

// This file does the rendering or referencing to the HTML page and directly interfaces with that page's javascript handler.
// Interfacing with the page's javascript event handler could also be done through the server side event routine (sse).

func Server() (*http.Server) {
	// Mux init
	themux := mux.NewRouter()
	source := &http.Server{Addr: ":8000", Handler:themux}		// web page created at localhost:8000 -> 127.0.0.1:8000 in this particular case.

	// Static Page Setup
	stat := http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/")))
	themux.PathPrefix("/static/").Handler(stat)

	// SSE Broker
	broker := sse.NewServer() 				// broker server init
	themux.Handle("/event", broker)   	// http event set
	// eventPoster(broker) 					// broker posting service  // hit the go func in loop for the broker

	// TODO: A go function and broker event pushing service loop is required.

	// Server GET AND POST
	themux.HandleFunc("/", mainpage).Methods("GET")
	http.Handle("/", themux)

	go func() {
		if err := source.ListenAndServe(); err != nil {
			tag.Fatal(fmt.Sprintf("ERROR - Socket not opened : %s", err))
		}
	}()

	return source
}

func render(w http.ResponseWriter, filename string, data []byte) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mainpage(res http.ResponseWriter, req *http.Request) {
	render(res, "web/template/mainpage.html", nil)
	//TODO: req is unused as of yet.
}