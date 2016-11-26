package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	projID  = "udy-demo"
	logName = "bucket-log"
	lggr    = &gLogger{}
)

func init() {
	var err error
	lggr, err = newClient(projID)
	if err != nil {
		panic(err)
	}
	makecert()
}

func gLog(msg string) {
	lggr.writeEntry(logName, msg)
}

func logReader(w io.Writer) {
	glog, err := newClient(projID)
	if err != nil {
		log.Printf("Failed to create logging client: %v\n", err)
		return
	}
	//log.Print("Fetching and printing log entries.")
	entries, err := glog.getEntries(projID, logName)
	if err != nil {
		log.Printf("Could not get entries: %v\n", err)
		return
	}
	//log.Printf("Found %d entries.", len(entries))
	for _, entry := range entries {
		fmt.Fprintf(w, "Entry: %6s @%s: %v\n",
			entry.Severity,
			entry.Timestamp.Format(time.RFC3339),
			entry.Payload)
	}
}

func logLine(msg string) {
	glog, err := newClient(projID)
	if err != nil {
		log.Printf("Failed to create logging client: %v\n", err)
	} else {
		glog.writeEntry(logName, msg)
	}
}

func logWrite(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	logLine(req.FormValue("msg"))
}

func readLogs(w http.ResponseWriter, req *http.Request) {
	logReader(w)
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	log.Println("hello")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func bucketEvent(w http.ResponseWriter, r *http.Request) {
	log.Println("bucket")
	if r.Method == "POST" {
		for k, v := range r.Header {
			if strings.HasPrefix(k, "X-Goog") {
				msg := k + " IS " + strings.Join(v, "")
				//fmt.Println(msg)
				gLog(msg)
			}
		}
	}
}

func main() {
	checkCert()
	addr := ":8443"
	s := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	http.HandleFunc("/hello", HelloServer)
	http.HandleFunc("/log/read", readLogs)
	http.HandleFunc("/log/write", logWrite)
	http.HandleFunc("/bucket/event", bucketEvent)
	log.Println("Start server --", addr)
	logLine("Start server -- " + addr)
	err := s.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
