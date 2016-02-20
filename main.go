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

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(s)
	//r.PathPrefix("/static").Handler(http.FileServer(http.Dir("./app/static/")))
	//fs := http.FileServer(http.Dir("app/static"))
	//r.Handle("/static/", http.StripPrefix("/static/", fs))
	//r.Handle("/static/", http.StripPrefix("/static/", fs))
	//r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/tmp"))))
	//r.Handle("/static", http.FileServer(http.Dir("app/static")))
	http.Handle("/", r)
	log.Println("Running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request to /")
	name := "app/static/html/index.html"
	http.ServeFile(w, r, name)
}

func ProjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a request to /project")
	//p, _ := loadPage(title)
	t, err := template.ParseFiles("app/templates/project.html.tmpl")
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
