package main

import (
	"cooleh/utils/ip"
	"cooleh/utils/levenshtein"
	"cooleh/utils/ogre"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	} else {
		// if there's no extension and the method is not GET, add the method to the filename
		// this allows users to specify the responses for POST, PUT, etc. as api/customers POST.json
		if r.Method != "" && r.Method != "GET" {
			fileName += " " + r.Method
		}

		// if there's no extension check if there's an html file with that name
		extension = ".html"
		_, err := os.ReadFile(fileName + extension)
		if err == nil {
			fileName += extension
			mimeType = mimetypes[extension]
		} else {
			// if there's no html file, check for json (to simulate an API get)
			extension = ".json"
			_, err = os.ReadFile(fileName + extension)
			if err == nil {
				fileName += extension
				mimeType = mimetypes[extension]
			}
		}
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
			fmt.Printf("\u001B[90mHint: create a file called '%v' to send a custom response for this request\u001B[0m\n", fileName+".json")

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

func server(port string) {
	networkIp := ip.Find()

	fmt.Printf("cooleh is running on http://127.0.0.1:%s or http://%s:%s\n", port, networkIp, port)

	http.HandleFunc("/", serveFile)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {

	if err := func() (rootCmd *cobra.Command) {

		var port string
		rootCmd = &cobra.Command{
			Use:   "cooleh",
			Short: "\u001B[1;38;2;116;132;116mcooleh\u001B[0m · ultra lightweight dev server · https://github.com/MattSimmons1/cooleh",
			Args:  cobra.ArbitraryArgs,
			Run: func(c *cobra.Command, args []string) {

				var wg sync.WaitGroup

				// we are going to wait for one goroutine to finish (but it never will)
				wg.Add(1)

				Ogre := ogre.New("o")

				Ogre.Growl()

				go server(port)

				wg.Wait()

				return
			},
		}
		rootCmd.Flags().StringVarP(&port, "port", "p", "5000", "change port")

		return
	}().Execute(); err != nil {
		panic(err)
	}
}
