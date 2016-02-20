package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello Steelhacks!")

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)
	log.Println("Running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request!")
	name := "app/static/html/index.html"
	http.ServeFile(w, r, name)
}
