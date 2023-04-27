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
	document = "<html><style>body{color:#282a2e;font-family:monospace;font-size:10px;}</style><body>" + document + "</body><html>"
	_, _ = w.Write([]byte(document))
}

func serveFile(w http.ResponseWriter, r *http.Request) {

	fileName := r.URL.Path
	fileName = strings.TrimPrefix(fileName, "/")

	t := time.Now().Format("3:04:05")

	if fileName == "" || strings.HasSuffix(fileName, "/") {
		fileName += "index.html"
		// if there's no index just try to serve site map
		_, err := os.ReadFile(fileName)
		if err != nil {

			fmt.Printf("\n%v\n", t)
			fmt.Printf("· %v - \033[91m404\033[0m\n", r.URL.Path)

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

	// if serving HTML print a blank line
	if extension == ".html" {
		fmt.Println("")

		fmt.Printf("%v\n", t)
		//fmt.Printf("%v · %v\n", t, r.RemoteAddr)
	} else {
		t = strings.Repeat(" ", len(t))
	}

	if err != nil {

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

		fmt.Printf("· %v - \033[91m404%v\033[0m\n", r.URL.Path, help)

		if extension == ".html" {
			serveDirectory(w, fileName)
		} else {
			w.WriteHeader(404)
		}
		return
	} else {

		if r.URL.Path == ("/" + fileName) {
			fmt.Printf("· /%v - 200\n", fileName)

		} else {
			fmt.Printf("· %v \033[90m-> /%v\033[0m - 200\n", r.URL.Path, fileName)

		}

	}

	document := string(b)

	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
