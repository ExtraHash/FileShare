package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/ExtraHash/p2p"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

type api struct {
	router     *mux.Router
	peerListMu *sync.Mutex
	readMu     *sync.Mutex
	messages   *chan []byte
	p2p        *p2p.DP2P
	db         *db
}

func (a *api) initialize(p2p *p2p.DP2P, db *db) {
	a.p2p = p2p
	a.db = db
	a.getRouter()
}

func fileExists(filename string) bool {
	_, configErr := os.Stat(filename)
	if os.IsNotExist(configErr) {
		return false
	}
	return true
}

// Run starts the server.
func (a *api) run() {
	port := 10188
	log.Print("HTTP Starting API on port " + strconv.Itoa(port) + ".")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port),
		handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "PATCH"}),
			handlers.AllowedOrigins([]string{"*"}))(a.router)))
}

func (a *api) getRouter() {
	// initialize router
	a.router = mux.NewRouter()
	a.router.Handle("/file", a.FileHandler()).Methods("POST")
	a.router.Handle("/file/{fileID}", a.FileHandler()).Methods("GET")
	a.router.Handle("/file", a.FileListHandler()).Methods("GET")
}

// GetIP from http request
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}

// FileListHandler handles the file list endpoint.
func (a *api) FileHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Print("HTTP", req.Method, req.URL, GetIP(req))

		switch req.Method {
		case "GET":
			vars := mux.Vars(req)
			id, err := uuid.FromString(vars["fileID"])
			if err != nil {
				res.WriteHeader(500)
				break
			}

			file := File{}
			a.db.db.Find(&file, "id = ?", id.String())

			if file.ID == uuid.Nil.String() {
				// file doesn't exist
				res.WriteHeader(500)
				break
			}

			fileB, err := ioutil.ReadFile(fileFolder + "/" + id.String())
			if err != nil {
				// file doesn't exist
				res.WriteHeader(500)
				break
			}
			res.Header().Set("Content-Disposition", "attachment; filename="+file.FileName)
			res.WriteHeader(200)
			res.Write(fileB)
		case "POST":
			file, handler, err := req.FormFile("file")
			if err != nil {
				// file doesn't exist
				res.WriteHeader(500)
				break
			}
			defer file.Close()

			fileID := uuid.NewV4()

			filePath := fileFolder + "/" + fileID.String()

			f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				// file doesn't exist
				res.WriteHeader(500)
				break
			}
			defer f.Close()

			io.Copy(f, file)

			fileBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				// file doesn't exist
				res.WriteHeader(500)
				break
			}

			newFile := File{
				ID:       fileID.String(),
				FileName: handler.Filename,
				Data:     fileBytes,
			}

			broadcastB, err := json.Marshal(newFile)
			a.db.db.Create(&newFile)

			a.p2p.Broadcast(broadcastB)
			res.WriteHeader(200)
		}

	})
}

// FileHandler handles the file endpoint.
func (a *api) FileListHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Print("HTTP", req.Method, req.URL, GetIP(req))

		fileList := []File{}
		a.db.db.Find(&fileList)

		fileB, err := json.Marshal(fileList)
		if err != nil {
			log.Print(err)
			res.WriteHeader(500)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(200)
		res.Write(fileB)
	})
}
