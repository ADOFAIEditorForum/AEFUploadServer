package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func index(writer http.ResponseWriter, _ *http.Request) {
	content, err := os.ReadFile("./web/form.html")
	if err != nil {
		println(err)
	}

	_, err2 := fmt.Fprintf(writer, string(content))
	if err2 != nil {
		return
	}
}

func mainScript(writer http.ResponseWriter, _ *http.Request) {
	content, err := os.ReadFile("./web/main.js")
	if err != nil {
		println(err)
	}

	_, err2 := fmt.Fprintf(writer, string(content))
	if err2 != nil {
		return
	}
}

func upload(writer http.ResponseWriter, req *http.Request) {
	b, _ := io.ReadAll(req.Body)

	now := time.Now()
	millis := now.UnixMilli()

	filename := fmt.Sprintf("./adofai%d.zip", millis)
	err := os.WriteFile(filename, b, 0644)
	go process(filename, millis)

	if err != nil {
		println(err)
	}

	_, err2 := fmt.Fprintf(writer, "Success")
	if err2 != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/main.js", mainScript)
	http.HandleFunc("/upload", upload)

	err := http.ListenAndServe("localhost:3676", nil)
	println(err)
}
