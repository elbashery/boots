package main

import (
    "fmt"
    "./SatsGo/goline"
    "./SatsGo/helper"
    "strings"
)

var tokens = ""

func main() {
	botsmid := []string{}
	for _, token := range strings.Split(tokens, "\n"){
		botsmid = append(botsmid, token[:33])
	}
	for _, token := range strings.Split(tokens, "\n"){
		bot := goline.NewLogin()
		bot.Login(token)
		for _, x := range botsmid {
			if !helper.InArray(bot.Friends, x) && x != bot.MID {
				bot.FindAndAddContactsByMid_thread(x)
			}
		}
		fmt.Println(bot.Friends)
	}
}