package main

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

const DISCORD_ID = "710885082100924416"

func rpc(address string) {
	err := client.Login(DISCORD_ID)
	if err != nil {
		panic(err)
	}

	time := time.Now()
	err = client.SetActivity(client.Activity{
		State:      address,
		Details:    "Playing MCBE on 1.12",
		LargeImage: "tedac",
		LargeText:  "TedacMC",
		SmallImage: "mc",
		SmallText:  "Minecraft 1.12 Support",
		Timestamps: &client.Timestamps{
			Start: &time,
		},
	})

	if err != nil {
		panic(err)
	}
}
