package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	body := strings.Join(os.Args[1:], " ")

	accountId := os.Getenv("TWILIO_ACCOUNT_ID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	sendTo := os.Getenv("TWILIO_SEND_TO_NUMBER")
	sendFrom := os.Getenv("TWILIO_SEND_FROM_NUMBER")
	smsUrl := "https://api.twilio.com/2010-04-01/Accounts/" + accountId + "/Messages.json"

	msgData := url.Values{}
	msgData.Set("To", sendTo)
	msgData.Set("From", sendFrom)
	msgData.Set("Body", body)
	msgDataReader := *strings.NewReader(msgData.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", smsUrl, &msgDataReader)
	req.SetBasicAuth(accountId, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
	}
}
