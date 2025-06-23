package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)


const urlToLambda = "https://fqiqsmzdqa5ylxd4r5uz5i2hpy0ejsin.lambda-url.ap-southeast-1.on.aws/"
type SendableJson struct{
	GBR					bool    `json:"GBR"`
	NonGBR				bool    `json:"NonGBR"`
	IoT 			    bool    `json:"IoT"` 
	AVRGaming		    bool    `json:"AVRGaming"`
	Healthcare          bool    `json:"Healthcare"`
	Industry40          bool    `json:"Industry40"`
	IoTDevices          bool    `json:"IoTDevices"`
	PublicSafety        bool    `json:"PublicSafety"`
	SmartCityHome       bool    `json:"SmartCityHome"`
	SmartTransport      bool    `json:"SmartTransport"`
	Smartphone          bool    `json:"Smartphone"`
	LTECategory         int8    `json:"LTECategory"`
	PacketLossRate      float64 `json:"PacketLossRate"`
	PacketDelay         float64 `json:"PacketDelay"`
	Timestamp           int8	`json:"Time"`
}

func TransfromRawDataToJSON(networkMetrics *NetworkMetrics) []byte {
	var jsonData SendableJson
	jsonData.GBR = networkMetrics.GBR
	jsonData.NonGBR = !networkMetrics.GBR
	jsonData.AVRGaming = networkMetrics.AVRGaming
	jsonData.Healthcare = networkMetrics.Healthcare
	jsonData.Industry40 = networkMetrics.Industry40
	jsonData.IoT = networkMetrics.IoT
	jsonData.IoTDevices = networkMetrics.IoTDevices
	jsonData.PublicSafety = networkMetrics.PublicSafety
	jsonData.SmartCityHome = networkMetrics.SmartCityHome
	jsonData.SmartTransport = networkMetrics.SmartTransport
	jsonData.Smartphone = networkMetrics.Smartphone
	jsonData.LTECategory = networkMetrics.LTECategory
	jsonData.PacketLossRate = networkMetrics.PacketLossRate
	jsonData.PacketDelay = networkMetrics.PacketDelay
	jsonData.Timestamp = networkMetrics.Timestamp

	jsonDataBytes, _ := json.Marshal(jsonData)

	return jsonDataBytes
}

// This is where integration part will take place

func SendJSONRequestToLambda(metrics *NetworkMetrics, a *App) {
	jsonData := TransfromRawDataToJSON(metrics)
	req, err := http.NewRequest("POST", urlToLambda, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	    return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return
	}
	// For now
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Update
	fmt.Println("Sent Data to Cloud\tReceived Response:")
	fmt.Println(string(body))
	a.showJS(false, "<div style='background: yellow;>'Sent Data to Cloud			Received Response:<br />" + string(body)+"</div>")
}