package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

type Files struct {
	l hclog.Logger
}

func NewFiles(l hclog.Logger) *Files {
	return &Files{l}
}

func (f *Files) UploadFile(rw http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(rw, "Could not download file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	f.l.Debug("Uploaded file: %+v\n", handler.Filename)
	f.l.Debug("File size: %+v\n", handler.Size)
	f.l.Debug("MIME Header: %+v\n", handler.Header)

	// Convert id to string
	vars := mux.Vars(r)
	id := vars["id"]
	idStr := fmt.Sprintf("%v", id)

	// Create directory according to id
	dirPath := "./images/" + idStr
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		http.Error(rw, "Could not create directory", http.StatusInternalServerError)
		return
	}

	tempFile, err := os.Create(dirPath + "/" + handler.Filename)
	if err != nil {
		http.Error(rw, "Could not save file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(rw, "Could not write to file", http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("Successfully Uploaded File\n"))
}

func (f *Files) GetFile(rw http.ResponseWriter, r *http.Request) {
	// Convert id to string
	vars := mux.Vars(r)
	id := vars["id"]
	idStr := fmt.Sprintf("%v", id)

	// Directory according to id
	dirPath := "./images/" + idStr

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		http.Error(rw, "Could not read directory", http.StatusInternalServerError)
		return
	}

	// Check if there are any files in the directory
	if len(files) == 0 {
		http.Error(rw, "No files in directory", http.StatusNotFound)
		return
	}

	// Serve the first file
	http.ServeFile(rw, r, dirPath+"/"+files[0].Name())
}
