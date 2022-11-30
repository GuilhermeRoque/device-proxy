package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"lorawanMgnt/application"
	"net/url"
	"os"
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
	appMgr *application.ApplicationMgr
}

func (wsClient *WsClient) Run() {
	wsClient.connect()
	wsClient.appMgr = &application.ApplicationMgr{}

	defer wsClient.close()
	for {
		messageType, message, err := wsClient.conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %s", err)
			return
		}
		log.Printf("Received message type %d payload %s", messageType, message)
		var newApplications []application.Application
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
	u := url.URL{Scheme: "ws", Host: os.Getenv("APPLICATION_MGR")}
	appURL := u.String()
	log.Printf("connecting to %s", appURL)

	conn, _, err := websocket.DefaultDialer.Dial(appURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	wsClient.conn = conn
}
func (wsClient *WsClient) close() {
	_ = wsClient.conn.Close()
}

func (wsClient *WsClient) parseBulkAddApplication(message []byte, newApplications *[]application.Application) error {
	return json.Unmarshal(message, &newApplications)
}

func (wsClient *WsClient) bulkAddApplication(newApplications []application.Application) *BulkAddApplicationResponse {
	var newApplicationsResponse []AddApplicationResponse
	var countOk int
	var countNok int
	var err error
	err = nil
	log.Printf("RECEIVED APPLICATIONS %+v", newApplications)
	for _, newApplication := range newApplications {
		isNew := wsClient.appMgr.IsApplicationNew(newApplication)
		if isNew {
			if os.Getenv("APP_ENV") != "local" {
				err = wsClient.appMgr.AddApplication(&newApplication)
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
				Message:       "Application already added",
			})
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
