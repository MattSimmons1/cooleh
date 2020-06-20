/*

  # cooleh

  A really simple webserver that serves any requested file as text/html.

  run with:

      export GOROOT=/usr/local/Cellar/go/1.14.3/libexec

      go build
      ./cooleh

  or add to your PATH and run:

      go install
      cooleh

*/


package main

import (
  "cooleh/utils/ip"
  "cooleh/utils/levenshtein"
  "cooleh/utils/ogre"
  "fmt"
  "io/ioutil"
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
  ".mjs":  "text/javascript; charset=utf-8",
  ".pdf":  "application/pdf",
  ".png":  "image/png",
  ".py":   "text/plain",
  ".svg":  "image/svg+xml",
  ".txt":   "text/plain",
  ".wasm": "application/wasm",
  ".webp": "image/webp",
  ".xml":  "text/xml; charset=utf-8",
}

func serveFile(w http.ResponseWriter, r *http.Request) {

  fileName := r.URL.Path
  fileName = strings.TrimPrefix(fileName, "/")

  t := time.Now().Format("3:04:05")

  if fileName == "" || strings.HasSuffix(fileName, "/") {
    fileName += "index.html"
    // if there's no index just try to serve site map
    _, err := ioutil.ReadFile(fileName)
    if err != nil {

      fmt.Printf("\n%v\n", t)
      fmt.Printf("· %v - \033[91m404\033[0m\n", r.URL.Path)

      var files []string

      err := filepath.Walk(".",
        func(path string, info os.FileInfo, err error) error {
          if err != nil {
            return err
          }

          if strings.HasSuffix(path, ".html") {
            files = append(files, "<a href=\"/" + path + "\">" + path + "</a>")
          }

          return nil
        })
      if err != nil {
        log.Println(err)
        return
      }

      document := strings.Join(files, "<br>")

      w.WriteHeader(200)
      document = "<p>404 - did you mean one of these:</p> </ br>" + document
      _, _ = w.Write([]byte(document))
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
    _, err := ioutil.ReadFile(fileName + ".html")
    if err == nil {
      extension = ".html"
      fileName += extension
      mimeType = mimetypes[extension]
    }
  }
  if mimeType == "" {
    mimeType = mimetypes["default"]
  }

  b, err := ioutil.ReadFile(fileName)


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
        if levenshtein.Distance(fileName, suggestion) < 5 {
          help = fmt.Sprintf(" - did you mean '%v'?", suggestion)

        }
    }

    fmt.Printf("· %v - \033[91m404%v\033[0m\n", r.URL.Path, help)

    w.WriteHeader(404)
    if extension == ".html" {
      w.Header().Set("Content-Type", mimeType)
      _, _ = w.Write([]byte("404 - not found"))
    }
    return
  } else{

    if r.URL.Path == ("/" + fileName) {
      fmt.Printf("· /%v - 200\n", fileName)

    } else {
      fmt.Printf("· %v \033[90m-> /%v\033[0m - 200\n", r.URL.Path, fileName)

    }

  }

  document := string(b)

  w.Header().Set("Content-Type", mimeType)
  w.WriteHeader(200)
  _, _ = w.Write([]byte(document))
}

func server() {
  networkIp := ip.Find()

  fmt.Printf("cooleh is running on http://127.0.0.1:5000 or http://%s:5000\n", networkIp)

  http.HandleFunc("/", serveFile)
  log.Fatal(http.ListenAndServe(":5000", nil))
}

func main() {

  //fmt.Printf("\033[90m • 89 · \033[0m\n")
  //fmt.Printf("\033[90m • 90 · \033[0m\n")
  //fmt.Printf("\033[91m • 91 · \033[0m\n")
  //fmt.Printf("\033[92m • 92 · \033[0m\n")
  //fmt.Printf("\033[93m • 93 · \033[0m\n")
  //fmt.Printf("\033[94m • 94 · \033[0m\n")
  //fmt.Printf("\033[95m • 95 · \033[0m\n")
  //fmt.Printf("\033[96m • 96 · \033[0m\n")
  //fmt.Printf("\033[97m • 97 · \033[0m\n")


  var wg sync.WaitGroup

  // we are going to wait for one goroutine to finish (but it never will)
  wg.Add(1)

  Ogre := ogre.New("o")

  Ogre.Growl()

  go server()

  wg.Wait()

}