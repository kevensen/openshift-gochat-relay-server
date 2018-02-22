//The entry point for the gochat program
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
)

var templatePath *string
var clientList map[string]time.Time

type host struct {
	name string
}

func refreshHostList(ticker time.Ticker) {
	for {
		select {
		case <-ticker.C:
			for client := range clientList {
				duration := int(time.Since(clientList[client]).Seconds())
				glog.Infoln("Seconds since host", client, "checked in", duration)
				if duration > 10 {
					delete(clientList, client)
					glog.Infoln("Dropping", client, "due to inactivity.")
				}

			}
		}
	}
}

//Primary handler
func ServeJSON(w http.ResponseWriter, r *http.Request) {

	client := r.RemoteAddr

	clientList[client] = time.Now()

	glog.Infoln("Request from host", client, "at", clientList[client])
	glog.Infoln(r.Header.Get("X-Forwarded-For"))

	keys := reflect.ValueOf(clientList).MapKeys()
	strkeys := make([]string, len(keys))

	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}

	js, err := json.Marshal(strkeys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func ClearHostList(w http.ResponseWriter, r *http.Request) {
	hostName := strings.Split(r.Host, ":")[0]

	if hostName == "localhost" {
		clientList = make(map[string]time.Time)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Unauthorized"))
	}
}

func main() {
	var host = flag.String("host", ":8080", "The host address of the application.")
	flag.Parse()
	ticker := time.NewTicker(time.Second * 5)

	clientList = make(map[string]time.Time)
	go refreshHostList(*ticker)
	http.HandleFunc("/", ServeJSON)
	http.HandleFunc("/clear", ClearHostList)

	fmt.Println("Starting the web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		glog.Fatalf("ListenAndServe: %s", err)
	}
}
