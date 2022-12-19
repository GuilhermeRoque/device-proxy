package application

type DeviceUplink struct {
	End_device_ids struct {
		Device_id       string
		Application_ids struct {
			Application_id string
		}
	}
	Received_at    string
	Uplink_message struct {
		Frm_payload string
	}
}

type DeviceJoin struct {
	End_device_ids struct {
		Device_id       string
		Application_ids struct {
			Application_id string
		}
	}
}

type DownlinkPayload struct {
	Frm_payload string `json:"frm_payload"`
	F_port      uint8  `json:"f_port"`
	Priority    string `json:"priority"`
}

type DownlinkMsg struct {
	Downlinks []DownlinkPayload `json:"downlinks"`
}
