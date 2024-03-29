package server

import (
	"fmt"
	"github.com/MattSimmons1/cooleh/utils/ip"
	"github.com/MattSimmons1/cooleh/utils/levenshtein"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var mimetypes = map[string]string{
	"default": "application/octet-stream",

	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".go":   "text/plain",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".ico":  "image/x-icon",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".js":   "text/javascript; charset=utf-8",
	".json": "application/json; charset=utf-8",
	".mjs":  "text/javascript; charset=utf-8",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".py":   "text/plain",
	".svg":  "image/svg+xml",
	".txt":  "text/plain",
	".wasm": "application/wasm",
	".webp": "image/webp",
	".xml":  "text/xml; charset=utf-8",
}

const dash = "\u001B[90m->\u001B[0m"

func serveDirectory(w http.ResponseWriter, filePath string) {
	var files []string

	// trim to get directory path
	i := strings.LastIndex(filePath, "/")
	if i >= 0 {
		filePath = filePath[:i]
	} else {
		filePath = "."
	}

	err := filepath.Walk(filePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".html") {
				files = append(files, "<a href=\"/"+path+"\">"+path+"</a>")
			}

			return nil
		})
	if err != nil {
		// it could be that the directory doesn't exist - attempt one directory down - recurse until we get to the root
		if filePath != "." {
			i = strings.LastIndex(filePath, "/")
			if i >= 0 {
				filePath = filePath[:i]
			} else {
				filePath = "."
			}
			serveDirectory(w, filePath)
		}
		return
	}

	document := strings.Join(files, "<br>")

	w.WriteHeader(200)
	w.Header().Set("Content-Type", mimetypes[".html"])
	document = "<p>404 - did you mean one of these:</p> </ br>" + document
	document = "<html><style>body{color:#282a2e;font-family:monospace;}</style><body>" + document + "</body><html>"
	_, _ = w.Write([]byte(document))
}

func serveFile(w http.ResponseWriter, r *http.Request) {

	// set CORS headers to 'allow all' for all requests
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	isGet := r.Method == "" || r.Method == "GET"

	fileName := r.URL.Path
	fileName = strings.TrimPrefix(fileName, "/")

	t := fmt.Sprintf("\u001B[90m%v\u001B[0m", time.Now().Format("3:04:05"))

	if isGet && (fileName == "" || strings.HasSuffix(fileName, "/")) {

		// check if we can redirect to path/index.html or /index.json
		if _, err := os.ReadFile(fileName + "index.html"); err == nil {
			fileName += "index.html"
		} else if _, err := os.ReadFile(fileName + "index.json"); err == nil {
			fileName += "index.json"
		} else {

			// if there's no index just try to serve site map
			fmt.Printf("\n")
			fmt.Printf(t+" %v %v \033[91m404\033[0m\n", r.URL.Path, dash)

			serveDirectory(w, fileName)
			return
		}
	}

	// pick a mime type based on the file extension
	mimeType := ""
	extension := ""
	i := strings.LastIndex(fileName, ".")
	if i > -1 {
		extension = fileName[i:]
		mimeType = mimetypes[extension]
	} else if _, err := os.ReadFile(fileName); err == nil {
		// if there's no extension and a file exists with that extension then assume it's text
		mimeType = mimetypes[".txt"]
	} else {
		// if there's no extension and the method is not GET, add the method to the filename
		// this allows users to specify the responses for POST, PUT, etc. as api/customers POST.json
		if r.Method != "" && r.Method != "GET" {
			fileName += " " + r.Method
		}

		// if there's no extension check if there's file with that name + a common extension
		extensions := []string{".html", ".json", ".csv", ".txt", ".xml"}
		for i := range extensions {
			extension = extensions[i]
			_, err := os.ReadFile(fileName + extension)
			if err == nil {
				fileName += extension
				mimeType = mimetypes[extension]
				break
			}
		}
		// if there are no matches, we will attempt to open a file with no extension and get an error
	}
	if mimeType == "" {
		mimeType = mimetypes["default"]
	}

	b, err := os.ReadFile(fileName)

	// if serving HTML print a blank line to signify a new session
	if extension == ".html" {
		fmt.Println("")
	}

	if err != nil {

		// if method isn't GET, send dummy responses
		if r.Method == "OPTIONS" {
			fmt.Printf("\n")
			fmt.Printf(t+" \u001B[94mOPTIONS\u001B[0m %v \u001B[90m-> None\u001B[0m %v 200\n", r.URL.Path, dash)
			w.WriteHeader(200)
			return
		} else if !isGet {
			fmt.Printf("\n")
			fmt.Printf(t+" \u001B[94m%v\u001B[0m %v \u001B[90m-> {}\u001B[0m %v 200\n", r.Method, r.URL.Path, dash)
			fmt.Printf("\u001B[90mHint: create a file called '%v' to send a custom response for this request\u001B[0m\n", fileName+".json")

			w.Header().Set("Content-Type", mimetypes[".json"])
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{}`))
			return
		}

		var files []string
		help := ""

		err := filepath.Walk(".",
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if path != "." && path != fileName {
					files = append(files, path)
				}

				return nil
			})
		if err != nil {
			log.Println(err)
			return
		}

		if len(files) > 0 {
			suggestion := levenshtein.Suggest(fileName, files)
			distance := levenshtein.Distance(fileName, suggestion)
			if distance < 5 && distance < len(fileName)/2 {
				help = fmt.Sprintf(" - did you mean '%v'?", suggestion)

			}
		}

		fmt.Printf(t+" %v - \033[91m404%v\033[0m\n", r.URL.Path, help)

		if extension == ".html" {
			serveDirectory(w, fileName)
		} else {
			w.WriteHeader(404)
		}
		return
	} else {

		if r.URL.Path == ("/" + fileName) {
			fmt.Printf(t+" /%v %v 200\n", fileName, dash)

		} else {
			fmt.Printf(t+" %v \033[90m-> /%v\033[0m %v 200\n", r.URL.Path, fileName, dash)
		}

	}

	document := string(b)

	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(200)
	_, _ = w.Write([]byte(document))
}

func Serve(port string) {
	networkIp := ip.Find()

	fmt.Printf("cooleh is running on http://127.0.0.1:%s or http://%s:%s\n", port, networkIp, port)

	http.HandleFunc("/", serveFile)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
