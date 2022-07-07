package main

import (
	"log"
	"lorawanMgnt/ttn"
	"net/http"
	"os"
	"strconv"
)

var brokerTTN = os.Getenv("brokerTTN")
var portTTN = os.Getenv("portTTN")
var passwordTTN = os.Getenv("passwordTTN")

var ttnClients []*ttn.TtnClient

func addApplication(applicationName string) {
	portttnInt, _ := strconv.Atoi(portTTN)
	ttnClient := ttn.NewTtnClient(brokerTTN, uint32(portttnInt), applicationName, passwordTTN)
	ttnClient.Connect()
	ttnClients = append(ttnClients, ttnClient)
}
func addApplicationHandler(w http.ResponseWriter, r *http.Request) {
	applicationName := r.FormValue("name")
	addApplication(applicationName)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	http.HandleFunc("/", addApplicationHandler)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
	//select {}
}
