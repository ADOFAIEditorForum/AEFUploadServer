package main

import (
	"fmt"
	"net/http"
)

func index(writer http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(writer, "Hello, World!")
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", index)

	err := http.ListenAndServe("localhost:3676", nil)
	println(err)
}
