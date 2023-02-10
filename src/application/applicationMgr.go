package application

import (
	"log"
	"os"
	"strconv"
)

type ApplicationMgr struct {
	ttnApplications []*TtnApplication
}

func (appMgr *ApplicationMgr) AddApplication(app *Application) error {
	var brokerTTN = os.Getenv("BROKER_TTN")
	var portTTN = os.Getenv("PORT_TTN")
	portTTNInt, _ := strconv.Atoi(portTTN)
	ttnApplication := NewTtnApp(brokerTTN, uint32(portTTNInt), app)
	log.Printf("Connecting to broker %s:%s to listen to app %s with API key %s\n", brokerTTN, portTTN, app.Name, app.APIKey)
	ttnApplication.Connect()
	appMgr.ttnApplications = append(appMgr.ttnApplications, ttnApplication)
	for _, device := range ttnApplication.app.Devices {
		if !device.Configured || (device.Service == ServiceCfg{}) {
			continue
		}
		ttnApplication.sendDeviceCfg(device)
		device.UpdateCfg = false

	}
	return nil
}

func (appMgr *ApplicationMgr) IsApplicationNew(newApplication Application) bool {
	isNew := true
	for _, ttnApp := range appMgr.ttnApplications {
		if ttnApp.app.ApplicationId == newApplication.ApplicationId {
			isNew = false
			break
		}
	}
	return isNew
}

func (appMgr *ApplicationMgr) UpdateConfigDevices(newApplication Application) bool {
	for _, ttnApp := range appMgr.ttnApplications {
		if ttnApp.app.ApplicationId == newApplication.ApplicationId {
			log.Println("Found application! Updating devices...")
			for index, device := range ttnApp.app.Devices {
				log.Printf("Looking for old device %v", device)
				for _, deviceNew := range newApplication.Devices {
					log.Printf("Looking for new device %v", deviceNew)
					if deviceNew.DeviceId == device.DeviceId {
						log.Printf("Found device!")
						if true {
							//if deviceNew.UpdateCfg {
							log.Printf("Device new will be updated")
							if !deviceNew.Configured || (deviceNew.Service == ServiceCfg{}) {
								log.Printf("Device new has no cfg")
							} else {
								log.Printf("Device new has cfg and will be updated")
								deviceNew.UpdateCfg = false
								ttnApp.app.Devices[index] = deviceNew
								ttnApp.sendDeviceCfg(deviceNew)
							}
						} else {
							log.Printf("Device new will not be updated %v", device)
						}

					}
				}
			}
			return true
		}
	}
	log.Println("Application not found...")
	return false
}
