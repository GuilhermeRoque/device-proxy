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
			for _, device := range ttnApp.app.Devices {
				if !device.Configured || (device.Service == ServiceCfg{}) {
					continue
				}
				for index, deviceNew := range newApplication.Devices {
					if deviceNew.DeviceId == device.DeviceId {
						ttnApp.app.Devices[index] = deviceNew
						ttnApp.sendDeviceCfg(deviceNew)
						break
					}
				}
			}
			break
		}
	}
	return false
}
