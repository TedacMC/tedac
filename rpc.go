package main

import (
	"time"

	"github.com/hugolgst/rich-go/client"
)

func rpc() {
	err := client.Login("DISCORD_ID")
	if err != nil {
		panic(err)
	}

	time := time.Now()
	err = client.SetActivity(client.Activity{
		State:      readConfig().Connection.RemoteAddress,
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
