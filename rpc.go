package main

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

const discordId = "710885082100924416"

func rpc(address string) {
	err := client.Login(discordId)
	if err != nil {
		panic(err)
	}

	t := time.Now()
	err = client.SetActivity(client.Activity{
		State:      address,
		Details:    "Playing Minecraft: Bedrock Edition on 1.12",
		LargeImage: "tedac",
		LargeText:  "TedacMC",
		SmallImage: "mc",
		SmallText:  "Minecraft 1.12 Support",
		Timestamps: &client.Timestamps{
			Start: &t,
		},
	})

	if err != nil {
		panic(err)
	}
}
