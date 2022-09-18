package main

import (
	"fmt"
	"time"

	"github.com/hugolgst/rich-go/client"
)

func rpc() {
	err := client.Login(DISCORD_ID)
	if err != nil {
		panic(err)
	}

	time := time.Now()
	err = client.SetActivity(client.Activity{
		State:      "TedacMC",
		Details:    fmt.Sprintf("Playing MCBE 1.12 on %s", readConfig().Connection.RemoteAddress),
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
