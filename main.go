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
	periodSec          = 60
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
	go PeriodUpdate(periodSec)
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
		if ct, ok := typeMap[suffix]; !ok {
			contentType = defaultContentType
		} else {
			contentType = ct
		}
	}
	var fStatus os.FileInfo
	var statusErr error
	if fStatus, statusErr = os.Stat(file); statusErr != nil {
		go deleteEntry(file)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	} else if fStatus.IsDir() {
		file = filepath.Join(file, index)
		if fStatus, statusErr = os.Stat(file); statusErr != nil {
			go deleteEntry(file)
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

func PeriodUpdate(peroidSec int) {
	ticker := time.NewTicker(time.Duration(int(time.Second) * peroidSec))
	for _ = range ticker.C {
		for file, record := range cacheFile {
			if fStatus, statusErr := os.Stat(file); statusErr != nil {
				deleteEntry(file)
			} else {
				if !fStatus.ModTime().Equal(record.modifiedTime) {
					go reload2Cache(fStatus.Size(), fStatus.ModTime(), file)
				}
			}
		}
	}
}

// delete the corresponding entry in cache
func deleteEntry(file string) {
	delete(cacheFile, file)
}

func reload2Cache(fSize int64, modifiedTime time.Time, file string) {
	fh, _ := os.Open(file)
	defer fh.Close()
	content := make([]byte, fSize)
	size, _ := fh.Read(content)
	cacheFile[file].modifiedTime = modifiedTime
	cacheFile[file].buf = content[:size]
}

func printUsage() {
	fmt.Println("Args: [port] [rootDir]")
	fmt.Println("Port should be a number, default value is 80")
	fmt.Println("RootDir should be an existed and absolute dir, default value is the working dir")
}
