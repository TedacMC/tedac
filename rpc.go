package main

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

// discordID represents the Discord application ID of Tedac.
const discordID = "710885082100924416"

// startRPC starts the Discord Rich Presence module of Tedac.
func (a *App) startRPC() {
	err := client.Login(discordID)
	if err != nil {
		return
	}

	start := time.Now()
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err = client.SetActivity(client.Activity{
				State:      a.remoteAddress,
				Details:    "Playing Minecraft: Bedrock Edition on 1.12",
				LargeImage: "tedac",
				LargeText:  "TedacMC",
				SmallImage: "mc",
				SmallText:  "Minecraft 1.12 Support",
				Timestamps: &client.Timestamps{
					Start: &start,
				},
			})
			if err != nil {
				return
			}
		case <-a.c:
			return
		}
	}
}
