
# cooleh

Ultra lightweight server written in go. Intended for quick testing of static sites and web experiments.

Created because `python -m http.server` is too hard to remember.

![](example.png)

Features:
- clear logging in the terminal for each request, separated by session
- sends the correct content type headers for all files, including .wasm
- tells you your IP address within the local network
- returns matching files when no file extension is provided, e.g. `/api/data` will return `api/data.json`
- catches typos and suggests corrections 

### Run

Download or build the binary and add it to your path.

In a directory with an HTML file, run:

    cooleh
    
### Build

To build the binary yourself and install:

    go install github.com/MattSimmons1/cooleh


### Import as go function

If you wanted to run the cooleh server from within a go project, you can import the server and run it like so:

```go
// main.go

package main

import "github.com/MattSimmons1/cooleh/server"

func main() {
  server.Serve("5000")
}
```

```shell
go run main.go
```