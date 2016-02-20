package main

import (
	"fmt"
	"html/template"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello Steelhacks!")

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/project", ProjectHandler)
	r.HandleFunc("/upload", UploadGetHandler).Methods("GET")
	r.HandleFunc("/upload", UploadGetHandler).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", r)
	log.Println("Running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request to /")
	name := "static/html/index.html"
	http.ServeFile(w, r, name)
}

func ProjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request to /project")
	//p, _ := loadPage(title)
	t, err := template.ParseFiles("templates/project.html.tmpl")
	if err != nil {
		log.Error(err)
	}
	t.Execute(w, nil)
}

func UploadGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request to /upload")
}

func UploadPostHandler(w http.ResponseWriter, r *http.Request) {

}
