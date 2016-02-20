package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	_ "net/http/httputil"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/dustinkirkland/golang-petname"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Hello Steelhacks!")

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/projects", ProjectHandler)
	r.HandleFunc("/upload", UploadGetHandler).Methods("GET")
	r.HandleFunc("/upload", UploadPostHandler).Methods("POST")
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

	Container.RLock()
	defer Container.RUnlock()
	log.Println("Received a request to /projects")

	projects := make([]*Project, 0)
	for _, p := range Container.projects {
		// add it to a slice of projects
		projects = append(projects, p)

	}

	t, err := template.ParseFiles("templates/project.html.tmpl")
	if err != nil {
		log.Error(err)
	}
	t.Execute(w, projects)
}

func UploadGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a GET request to /upload")
	name := "static/html/upload.html"
	http.ServeFile(w, r, name)
}

func UploadPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a POST request to /upload")

	file, header, err := r.FormFile("application")
	if err != nil {
		log.Error(err)
	}

	defer file.Close()
	fmt.Fprintf(w, "%v", header.Header)
	f, err := os.OpenFile("./applications/"+header.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error(err)
	}
	defer f.Close()
	io.Copy(f, file)

	// Container.AddProject(r)
	// TODO get the zip attachment
	// TODO redirect you to a success page!
	// Fire off a goroute to spin up the docker file
	// Add a file pointer into the struct
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Identifier  string `json:"identifier"`
	HostName    string `json:"host_name"`
}

type ProjectContainer struct {
	projects map[string]*Project
	sync.RWMutex
}

func (projCntr *ProjectContainer) AddProject(req *http.Request) {
	projCntr.Lock()
	defer projCntr.Unlock()
	decoder := json.NewDecoder(req.Body)
	proj := Project{}
	err := decoder.Decode(&proj)
	if err != nil {
		log.Error("Failed to unmarshal! Reproducing err and json.")
		log.Error(err)
	}

	host := projCntr.GenerateNewHostname()
	projCntr.projects[host] = &proj
}

// TODO GenerateNewHostname
func (proj *ProjectContainer) GenerateNewHostname() string {
	return "foobar"
}

var Container = &ProjectContainer{
	projects: make(map[string]*Project),
}

// TODO add the method that fires off the docker container.
// TODO add the method that adds the hostname to the router.
// TODO reverse proxy
