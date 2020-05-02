package main

import (
	"flag"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	//"time"
)

func main() {
	var (
		mode   = flag.String("mode", "demographics", "mode of delevry helper [multycast|reply|push]")
		date   = flag.String("date", "20200406", "date the message were sent , format 'yyymmdd'")
		client = &http.Client{}
	)

	flag.Parse()
	bot, err := linebot.New(
		"d804349249472af3567cff7f8ec5a179",
		"PGxXex43ijl2VMKtAYY0CbeShmgSiLj/WLjSfZ0oiDOq7E+NyR8ty4dIhSLXE2VpfMOlIxyCwW5APZAAzzPX2dOmWObndSJ4ypC661FLjmtRBKfSOoMpePargvVB5UFeO5mFxcD7SZ9p5bsToo9VSAdB04t89/1O/w1cDnyilFU=",
		linebot.WithHTTPClient(client),
	)
	if err != nil {
		log.Fatalf("authentication err is : %v", err)
	}
	switch *mode {
	case "messages":
		res, err := bot.GetNumberMessagesDelivery(*date).Do()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v", res)
	case "followers":
		res, err := bot.GetNumberFollowers(*date).Do()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v", res)
	case "demographics":
		res, err := bot.GetFriendDemographics().Do()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%v", res)

	default:
		log.Fatal("implement me")
	}
	result, err := bot.PushMessage("qwerty09.1", linebot.NewTextMessage("hello")).Do()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
}

/*	var res *linebot.MessagesNumberResponse
	switch *mode {
	case "multicast":
		res, err = bot.GetNumberMulticastMessages(*date).Do()
	case "push":
		res, err = bot.GetNumberPushMessages(*date).Do()
	case "reply":
		res, err = bot.GetNumberReplyMessages(*date).Do()
	case "broadcast":
		res, err = bot.GetNumberBroadcastMessages(*date).Do()
	default:
		log.Fatalf("implement me")
	}
	if err != nil {
		log.Println(err)
	}
	log.Printf("the resutl is : %v", res)
}
*/
