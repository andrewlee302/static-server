package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileRecord struct {
	buf          []byte
	modifiedTime time.Time
}

const (
	index              = "index.html"
	defaultContentType = "text/plain"
)

var (
	typeMap = map[string]string{
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"jpg":  "image/jpeg",
		"svg":  "text/xml"}
)

var (
	rootPath  string
	cacheFile map[string]*FileRecord
	port      int
)

func main() {
	if len(os.Args) > 1 {
		var err error
		if port, err = strconv.Atoi(os.Args[1]); err != nil {
			printUsage()
			return
		} else {
			if len(os.Args) > 2 {
				rootPath = os.Args[2]
				if rootStat, err := os.Stat(rootPath); err != nil || !rootStat.IsDir() {
					printUsage()
					return
				}
			} else {
				rootPath, _ = os.Getwd()
			}
		}
	} else {
		port = 80
		rootPath, _ = os.Getwd()
	}
	cacheFile = make(map[string]*FileRecord)
	server := http.NewServeMux()
	server.HandleFunc("/", service)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), server); err != nil {
		fmt.Println(err)
	}
}

func service(w http.ResponseWriter, req *http.Request) {
	var file, contentType string
	file = filepath.Clean(rootPath + filepath.Clean(req.RequestURI))
	idx := strings.Index(file, "?")
	if idx != -1 {
		file = file[:idx]
	}
	suffixIdx := strings.LastIndex(file, ".")
	if suffixIdx != -1 && suffixIdx < len(file)-1 {
		suffix := file[suffixIdx+1:]
		fmt.Println(suffix)
		if ct, ok := typeMap[suffix]; !ok {
			contentType = defaultContentType
		} else {
			contentType = ct
		}
	}
	var fStatus os.FileInfo
	var statusErr error
	if fStatus, statusErr = os.Stat(file); statusErr != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	} else if fStatus.IsDir() {
		file = filepath.Join(file, index)
		if fStatus, statusErr = os.Stat(file); statusErr != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404"))
			return
		}
	}
	if record, ok := cacheFile[file]; !ok {
		cacheFile[file] = new(FileRecord)
		reload2Cache(fStatus.Size(), fStatus.ModTime(), file)
	} else {
		// reloading is necessarily needed when modified time is not equal
		if !fStatus.ModTime().Equal(record.modifiedTime) {
			go reload2Cache(fStatus.Size(), fStatus.ModTime(), file)
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(cacheFile[file].buf)

}

func reload2Cache(fSize int64, modifiedTime time.Time, file string) {
	fh, _ := os.Open(file)
	content := make([]byte, fSize)
	size, _ := fh.Read(content)
	cacheFile[file].buf = content[:size]
	cacheFile[file].modifiedTime = modifiedTime
}

func printUsage() {
	fmt.Println("Args: [port] [rootDir]")
	fmt.Println("Port should be a number, default value is 80")
	fmt.Println("RootDir should be an existed and absolute dir, default value is the working dir")
}
