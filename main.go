package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func errorHandler(writer http.ResponseWriter, status int) {
	writer.WriteHeader(status)
	if status == http.StatusNotFound {
		_, err := fmt.Fprint(writer, "404 Not Found")
		if err != nil {
			println(err)
		}
	}
}

func index(writer http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		errorHandler(writer, http.StatusNotFound)
		return
	}

	switch req.Method {
	case "GET":
		content, err := os.ReadFile("./web/form.html")
		if err != nil {
			println(err)
		}

		_, err2 := fmt.Fprintf(writer, string(content))
		if err2 != nil {
			return
		}
	}
}

func upload(writer http.ResponseWriter, req *http.Request) {
	path := strings.Split(req.URL.Path, "/")
	if len(path) != 3 {
		errorHandler(writer, http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(path[2], 10, 64)
	if err != nil {
		errorHandler(writer, http.StatusBadRequest)
		return
	}

	var exists bool
	if _, exists = uploadSession.timeMap.data[sessionID]; !exists {
		errorHandler(writer, http.StatusNotFound)
		return
	}

	switch req.Method {
	case "POST":
		b, _ := io.ReadAll(req.Body)

		uploadSession.timeMap.data[sessionID] = SessionTimePair{time.Now().UnixMilli(), uploadSession.timeMap.data[sessionID].isCompleted}
		uploadSession.dataMap.data[sessionID] = append(uploadSession.dataMap.data[sessionID], b...)
		if len(uploadSession.dataMap.data[sessionID]) > (1024*1024*1024)*3/2 { // 1.5GiB
			_, err := fmt.Fprintf(writer, "Upload Failed: Too Large")
			if err != nil {
				return
			}

			delete(uploadSession.dataMap.data, sessionID)
			errorHandler(writer, http.StatusMethodNotAllowed)
		}

		_, err := fmt.Fprintf(writer, "Success")
		if err != nil {
			return
		}
	case "DELETE":
		b, exists := uploadSession.dataMap.data[sessionID]
		if !exists {
			errorHandler(writer, http.StatusNotFound)
			return
		}

		now := time.Now()
		millis := now.UnixMilli()

		filename := fmt.Sprintf("./adofai%d.zip", millis)
		err := os.WriteFile(filename, b, 0644)

		downloadID := make(chan uint64)
		go process(filename, millis, downloadID)

		if err != nil {
			println(err)
		}

		sessionIDMap.data[sessionID] = <-downloadID

		uploadSession.timeMap.data[sessionID] = SessionTimePair{uploadSession.timeMap.data[sessionID].latestUpload, true}
		delete(uploadSession.dataMap.data, sessionID)

		_, err2 := fmt.Fprintf(writer, "Upload Success")
		if err2 != nil {
			return
		}
	}
}

func mainScript(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		content, err := os.ReadFile("./web/main.js")
		if err != nil {
			println(err)
		}

		_, err2 := fmt.Fprintf(writer, string(content))
		if err2 != nil {
			return
		}
	}
}

type ByteArray []byte
type SessionMapPair struct {
	timeMap SessionTimeMap
	dataMap SessionDataMap
}

type SessionTimePair struct {
	latestUpload int64
	isCompleted  bool
}

type SessionTimeMap struct {
	data map[int64]SessionTimePair
}

type SessionIDMap struct {
	data map[int64]uint64
}

type SessionDataMap struct {
	data map[int64]ByteArray
}

var sessionIDMap = &SessionIDMap{
	data: make(map[int64]uint64),
}

var uploadSession = &SessionMapPair{
	SessionTimeMap{
		data: make(map[int64]SessionTimePair),
	},
	SessionDataMap{
		data: make(map[int64]ByteArray),
	},
}

func timeoutHandler(sessionID int64, timeoutDuration int64) {
	for {
		time.Sleep(1 * time.Millisecond)
		timeData, exists := uploadSession.timeMap.data[sessionID]
		if !exists || timeData.isCompleted || timeData.latestUpload <= time.Now().UnixMilli()-timeoutDuration {
			break
		}
	}

	fmt.Printf("Upload Session %d Closed\n", sessionID)
	delete(uploadSession.timeMap.data, sessionID)
	delete(uploadSession.dataMap.data, sessionID)
}

var chunkSize = 1024 * 1024 * 5

func getSession(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		var sessionID int64
		for {
			sessionID = rand.Int64()
			_, exists := uploadSession.timeMap.data[sessionID]
			if !exists {
				break
			}
		}

		fmt.Printf("Upload Session %d Opened\n", sessionID)
		uploadSession.timeMap.data[sessionID] = SessionTimePair{time.Now().UnixMilli(), uploadSession.timeMap.data[sessionID].isCompleted}
		uploadSession.dataMap.data[sessionID] = ByteArray{}

		_, err := fmt.Fprintf(writer, "{\"sessionID\":\"%d\", \"chunkSize\": %d}", sessionID, chunkSize)
		if err != nil {
			return
		}

		go timeoutHandler(sessionID, 1000*10)
	}
}

func getDownload(writer http.ResponseWriter, req *http.Request) {
	path := strings.Split(req.URL.Path, "/")
	if len(path) != 3 {
		errorHandler(writer, http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(path[2], 10, 64)
	if err != nil {
		errorHandler(writer, http.StatusBadRequest)
		return
	}

	fmt.Println(sessionID)

	var exists bool
	if _, exists = sessionIDMap.data[sessionID]; !exists {
		errorHandler(writer, http.StatusNotFound)
		return
	}

	switch req.Method {
	case "GET":
		_, err := fmt.Fprintf(writer, "%d", sessionIDMap.data[sessionID])
		if err != nil {
			return
		}
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/get_session", getSession)
	http.HandleFunc("/get_download/", getDownload)
	http.HandleFunc("/upload/", upload)

	http.HandleFunc("/main.js", mainScript)

	err := http.ListenAndServe("localhost:3676", nil)
	println(err)

	/*
		data, err := os.ReadFile("src/test.json")
		if err != nil {
			log.Fatal(err)
		}

		println(convertToValidJSON(string(data)))*/
}
