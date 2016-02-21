package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/dustinkirkland/golang-petname"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

var r *mux.Router

const (
	IP = "192.168.99.100"
)

func main() {
	fmt.Println("Hello Steelhacks!")

	r = mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Host("localhost")
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
	log.Println("Received a request to /projects")

	Container.RLock()
	defer Container.RUnlock()

	t, err := template.ParseFiles("templates/project.html.tmpl")
	if err != nil {
		log.Error(err)
	}
	t.Execute(w, Container.Projects)
}

func UploadGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a GET request to /upload")
	name := "static/html/upload.html"
	http.ServeFile(w, r, name)
}

func UploadPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a POST request to /upload")

	file, _, err := r.FormFile("application")
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	project := GetProjectFromRequest(r)
	Container.AddProject(project)

	f, err := os.OpenFile("./applications/"+project.Identifier, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error(err)
	}

	defer f.Close()
	io.Copy(f, file)

	if err := MakeDocker(project); err != nil {
		http.Error(w, "Failed to shit out the docker container", http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/success.html.tmpl")
	if err != nil {
		log.Error(err)
	}
	t.Execute(w, project)
}

func GetProjectFromRequest(r *http.Request) *Project {
	return &Project{
		Authors:     r.FormValue("authors"),
		Description: r.FormValue("description"),
		Title:       r.FormValue("title"),
		Identifier:  GenerateNewHostname(),
	}
}

type Project struct {
	Authors     string `json:"authors"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Identifier  string `json:"identifier"`
	Port        int
}

type ProjectContainer struct {
	Projects []*Project
	sync.RWMutex
	PortCounter int
}

func (projCntr *ProjectContainer) AddProject(project *Project) {
	projCntr.Lock()
	defer projCntr.Unlock()
	project.Port = projCntr.PortCounter
	projCntr.Projects = append(projCntr.Projects, project)
	subdomain := fmt.Sprintf("{subdomain:%s}", project.Identifier)
	r.Handle("/", ReverseProxy(project)).Host(subdomain)
	projCntr.PortCounter++
}

func ReverseProxy(project *Project) http.Handler {

	urlP := fmt.Sprintf("http://%v:%v", IP, project.Port)
	u, _ := url.Parse(urlP)
	return httputil.NewSingleHostReverseProxy(u)
}

func GenerateNewHostname() string {
	return petname.Generate(3, "-")
}

var Container = &ProjectContainer{
	Projects:    make([]*Project, 0),
	PortCounter: 9000,
}

func GetClient() *docker.Client {
	c, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func TryContainer(appPath string, port int, appName string) {
	client := GetClient()

	log.Info("Hello!")

	BuildImage(client, appPath, appName)
	LaunchContainer(client, port, BuildContainer(client, appName))
}

func BuildImage(client *docker.Client, dirpath, projectName string) {
	curr, _ := os.Getwd()
	appPath := path.Join(curr, dirpath)

	buffer := bytes.NewBuffer(nil)
	imageOps := docker.BuildImageOptions{
		Name:         projectName,
		OutputStream: buffer,
		ContextDir:   appPath,
	}

	if err := client.BuildImage(imageOps); err != nil {
		log.Fatal(err)
	}

	buffer.WriteTo(os.Stdout)
}

func GetImageConfig(client *docker.Client, imageTag string) *docker.Config {
	img, err := client.InspectImage(imageTag)
	if err != nil {
		log.Fatal(err)
	}
	//img.Config.ExposedPorts
	return img.Config
}

func BuildContainer(client *docker.Client, imageName string) *docker.Container {

	config := GetImageConfig(client, imageName)

	containerOps := docker.CreateContainerOptions{
		Name:   imageName,
		Config: config,
	}

	container, err := client.CreateContainer(containerOps)
	if err != nil {
		log.Error(err)
	}
	return container
}

func LaunchContainer(client *docker.Client, port int, container *docker.Container) {
	ports := make(map[docker.Port][]docker.PortBinding)
	portStr := docker.Port(fmt.Sprintf("%d/tcp", 8000))

	ports[portStr] = []docker.PortBinding{
		{
			HostPort: fmt.Sprintf("%d", port),
			HostIP:   "0.0.0.0",
		},
	}

	err := client.StartContainer(container.ID, &docker.HostConfig{
		PortBindings: ports,
		//	PublishAllPorts: true,
	})
	if err != nil {
		log.Error(err)
	}
}

// TODO add the method that adds the hostname to the router.
// TODO reverse proxy
func MakeDocker(project *Project) error {

	// first, get the zip file and unzip it.
	zipath := path.Join("applications/", project.Identifier)

	targetPath := path.Join("unzip/", project.Identifier)
	log.Info(zipath)
	log.Info(targetPath)

	err := unzip(zipath, targetPath)
	if err != nil {
		log.Info(err)
		return err
	}

	// then, get the path to the unzipped file
	// plug the path into the TryContainer call
	TryContainer(targetPath, project.Port, project.Identifier)

	return nil
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	log.Info(len(reader.File))
	for _, file := range reader.File {
		log.Info(file.Name)
		path := path.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			log.Error(err)
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Error(err)
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}
