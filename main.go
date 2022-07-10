package main

import (
	"log"
	"lorawanMgnt/ttn"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var brokerTTN = os.Getenv("BROKER_TTN")
var portTTN = os.Getenv("PORT_TTN")
var passwordTTN = os.Getenv("PASSWORD_TTN")

type TtnApplication struct {
	Name      string
	TtnClient *ttn.TtnClient
}

var ttnApplications []*TtnApplication

func parseStringList(stringList string) []string {
	stringList = strings.TrimPrefix(stringList, "[")
	stringList = strings.TrimSuffix(stringList, "]")
	return strings.Split(stringList, ", ")

}
func addApplication(applicationName string) {
	portttnInt, _ := strconv.Atoi(portTTN)
	ttnClient := ttn.NewTtnClient(brokerTTN, uint32(portttnInt), applicationName, passwordTTN)
	ttnClient.Connect()
	ttnApplications = append(ttnApplications, &TtnApplication{
		Name:      applicationName,
		TtnClient: ttnClient,
	})
}
func addApplicationHandler(w http.ResponseWriter, r *http.Request) {
	applicationNames := r.FormValue("name")
	log.Printf("RECEIVED APPLICATIONS %s", applicationNames)
	applicationNamesSlice := parseStringList(applicationNames)
	for _, applicationName := range applicationNamesSlice {
		isNew := false
		for _, application := range ttnApplications {
			if application.Name == applicationName {
				isNew = true
				break
			}
		}
		if isNew {
			addApplication(applicationName)
		}
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	//f := 32.1
	//var buf [4]byte
	//binary.BigEndian.PutUint32(buf[:], math.Float32bits(float32(f)))
	//fmt.Println(buf)
	//bits := binary.BigEndian.Uint32(buf[:])
	//fmt.Println(math.Float32frombits(bits))
	log.Println("Running...")
	http.HandleFunc("/", addApplicationHandler)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
	//select {}
}
