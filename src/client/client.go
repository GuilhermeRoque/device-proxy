package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	application2 "lorawanMgnt/src/application"
	"net/url"
	"os"
	"time"
)

type BulkAddApplicationResponse struct {
	Message  string                   `json:"message"`
	CountOK  int                      `json:"count_ok"`
	CountNOK int                      `json:"count_nok"`
	Details  []AddApplicationResponse `json:"details"`
}

type AddApplicationResponse struct {
	ApplicationId string `json:"applicationId"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
}

type WsClient struct {
	conn   *websocket.Conn
	appMgr *application2.ApplicationMgr
}

func (wsClient *WsClient) Run() {
	wsClient.connect()
	wsClient.appMgr = &application2.ApplicationMgr{}

	defer wsClient.close()
	for {
		messageType, message, err := wsClient.conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %s", err)
			return
		}
		log.Printf("Received message type %d payload %s", messageType, message)
		var newApplications []application2.Application
		err = wsClient.parseBulkAddApplication(message, &newApplications)
		if err != nil {
			log.Printf("Error parsing request %s", err)
			continue
		}
		response := wsClient.bulkAddApplication(newApplications)
		err = wsClient.conn.WriteJSON(response)
		if err != nil {
			log.Printf("Could not send report %s", err)
		}
	}
}

func (wsClient *WsClient) connect() {
	u := url.URL{Scheme: "ws", Host: os.Getenv("APP_MGR")}
	appURL := u.String()
	log.Printf("connecting to %s", appURL)

	for {
		conn, _, err := websocket.DefaultDialer.Dial(appURL, nil)
		if err != nil {
			log.Println("Error connection WS dial:", err)
			log.Println("Trying WS connection into 3 seconds...")
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("WS connected!")
		wsClient.conn = conn
		break
	}
}
func (wsClient *WsClient) close() {
	_ = wsClient.conn.Close()
}

func (wsClient *WsClient) parseBulkAddApplication(message []byte, newApplications *[]application2.Application) error {
	return json.Unmarshal(message, &newApplications)
}

func (wsClient *WsClient) bulkAddApplication(newApplications []application2.Application) *BulkAddApplicationResponse {
	var newApplicationsResponse []AddApplicationResponse
	var countOk int
	var countNok int
	var err error
	err = nil
	log.Printf("RECEIVED APPLICATIONS %+v", newApplications)
	for _, newApplication := range newApplications {
		log.Printf("Checking application %+v", newApplication)
		isNew := wsClient.appMgr.IsApplicationNew(newApplication)
		log.Printf("isNew %t", isNew)
		if isNew {
			if os.Getenv("APP_ENV") != "local" {
				err = wsClient.appMgr.AddApplication(newApplication)
			}
			if err != nil {
				newApplicationsResponse = append(newApplicationsResponse, AddApplicationResponse{
					ApplicationId: newApplication.ApplicationId,
					Status:        500,
					Message:       fmt.Sprintf("Could not add application due to error %s", err),
				})
			} else {
				newApplicationsResponse = append(newApplicationsResponse, AddApplicationResponse{
					ApplicationId: newApplication.ApplicationId,
					Status:        200,
					Message:       "Application added successfully",
				})
			}
		} else {
			newApplicationsResponse = append(newApplicationsResponse, AddApplicationResponse{
				ApplicationId: newApplication.ApplicationId,
				Status:        202,
				Message:       "Application already added. Devices cfg will be updated",
			})
			wsClient.appMgr.UpdateConfigDevices(newApplication)
		}
		countOk += 1
	}
	return &BulkAddApplicationResponse{
		Message:  "Added applications report",
		CountOK:  countOk,
		CountNOK: countNok,
		Details:  newApplicationsResponse,
	}
}
