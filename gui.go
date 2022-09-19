package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// Define the progress variables, a channel and a variable
var progressIncrementer chan bool
var progress float32

func run() {
	// Setup a separate channel to provide ticks to increment progress
	progressIncrementer = make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second / 25)
			progressIncrementer <- true
		}
	}()

	go func() {
		// create new window
		w := app.NewWindow(
			app.Title("Tedac: The 1.12 MCBE Proxy"),
			app.Size(unit.Dp(360), unit.Dp(360)),
			app.MaxSize(unit.Dp(360), unit.Dp(360)),
		)
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

type C = layout.Context
type D = layout.Dimensions

func draw(w *app.Window) error {
	// ops are the operations from the UI
	var ops op.Ops

	// startButton is a clickable widget
	var startButton widget.Clickable

	// boilDurationInput is a textfield to input boil duration
	var input widget.Editor

	var started bool
	//var address string

	// th defines the material design style
	th := material.NewTheme(gofont.Collection())

	for {
		select {
		// listen for events in the window.
		case e := <-w.Events():

			// detect what type of event
			switch e := e.(type) {
			// this is sent when the application should re-render.
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				// Let's try out the flexbox layout concept
				if startButton.Clicked() {
					// Start (or stop) the boil
					started = !started

					if started {
						addr := input.Text()
						fmt.Printf("Connecting... %s", addr)
						// go func() {
						// 	err := server("0.0.0.0:19132", addr)
						// 	if err != nil {
						// 		fmt.Println(err)
						// 	}
						// }()
					}
				}

				layout.Flex{
					// Vertical alignment, from top to bottom
					Axis: layout.Vertical,
					// Empty space is left at the start, i.e. at the top
					Spacing: layout.SpaceStart,
				}.Layout(gtx,
					layout.Rigid(
						func(gtx C) D {
							l := material.H1(th, "TedacMC")
							l.Alignment = text.Middle
							l.TextSize = 30
							return l.Layout(gtx)
						},
					),
					// The inputbox
					layout.Rigid(
						func(gtx C) D {
							// Wrap the editor in material design
							ed := material.Editor(th, &input, "Server Address")

							// Define characteristics of the input box
							input.SingleLine = true
							input.Alignment = text.Middle

							// ... and borders ...
							border := widget.Border{
								Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								CornerRadius: unit.Dp(3),
								Width:        unit.Dp(2),
							}
							// ... before laying it out, one inside the other
							return border.Layout(gtx, ed.Layout)

						},
					),

					// The button
					layout.Rigid(
						func(gtx C) D {
							// We start by defining a set of margins
							margins := layout.Inset{
								Top:    unit.Dp(25),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(10),
								Left:   unit.Dp(10),
							}
							// Then we lay out within those margins
							return margins.Layout(gtx,
								func(gtx C) D {
									// The text on the button depends on program state
									var text string
									if started {
										text = "Stop"
									} else {
										text = "Start"
									}
									btn := material.Button(th, &startButton, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
				)
				e.Frame(gtx.Ops)

			// this is sent when the application is closed.
			case system.DestroyEvent:
				return e.Err
			}

			// listen for events in the incrementor channel
		case <-progressIncrementer:
		}
	}
}
