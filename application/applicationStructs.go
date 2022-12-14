package application

type Application struct {
	Name          string   `json:"name"`
	APIKey        string   `json:"apiKey"`
	ApplicationId string   `json:"applicationId"`
	Bucket        string   `json:"bucket"`
	Token         string   `json:"token"`
	Devices       []Device `json:"devices"`
}

type Device struct {
	Name       string     `json:"name"`
	DeviceId   string     `json:"devId"`
	DeviceEUI  string     `json:"devEUI"`
	Service    ServiceCfg `json:"serviceProfile"`
	Configured bool       `json:"configured"`
}

type ServiceCfg struct {
	Name         string `json:"name"`
	DataType     uint8  `json:"dataType"`
	ChannelType  uint8  `json:"channelType"`
	ChannelParam uint8  `json:"channelParam"`
	Acquisition  uint8  `json:"acquisition"`
	Period       uint32 `json:"period"`
}
