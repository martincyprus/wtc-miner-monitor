// telegrambot.go
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	TelegramAPI = "https://api.telegram.org/"
)

func nodeIsDownTelegram(row HashlogItem, configuration Configuration) {
	telegramCall := TelegramAPI + configuration.TelegramBotAPIKey + "/sendMessage?chat_id=" + configuration.TelegramChannelID + "&parse_mode=markdown&text=*NODE+DOWN* nodeid: " + strconv.Itoa(row.Nodeid) + " nodename: " + row.Nodename + " last seen: " + row.Ts.UTC().Format("20060102+15:04+UTC")
	if strings.ToUpper(configuration.Debug) == "YES" {
		fmt.Println(telegramCall)
	}
	resp, _ := http.Get(telegramCall)
	resp.Body.Close()
}

func zeroPeersTelegram(row HashlogItem, configuration Configuration) {
	telegramCall := TelegramAPI + configuration.TelegramBotAPIKey + "/sendMessage?chat_id=" + configuration.TelegramChannelID + "&parse_mode=markdown&text=*Zero Peers for:* nodeid: " + strconv.Itoa(row.Nodeid) + " nodename: " + row.Nodename + "peer count: " + strconv.Itoa(row.Peercount) + "last seen: " + row.Ts.UTC().Format("20060102+15:04+UTC")
	if strings.ToUpper(configuration.Debug) == "YES" {
		fmt.Println(telegramCall)
	}
	resp, _ := http.Get(telegramCall)
	resp.Body.Close()
}

func testMyTelegramBot(configuration Configuration) {
	telegramCall := TelegramAPI + configuration.TelegramBotAPIKey + "/sendMessage?chat_id=" + configuration.TelegramChannelID + "&parse_mode=markdown&text=If you see this message it means that the bot connection is working for the wtc-miner-monitor"
	fmt.Println(telegramCall)
	resp, _ := http.Get(telegramCall)
	resp.Body.Close()
}
