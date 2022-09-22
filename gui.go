package main

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/gesture"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
	"github.com/inkeliz/giohyperlink"
)

type C = layout.Context
type D = layout.Dimensions

func run(w *app.Window) error {
	theme := material.NewTheme(gofont.Collection())

	var (
		ops         op.Ops
		startButton widget.Clickable
		richText    richtext.InteractiveText
	)
	address := widget.Editor{
		Alignment:  text.Middle,
		SingleLine: true,
		Submit:     true,
	}
	port := address
	port.Filter = "0123456789"
	startText := "Start"

	//fontSize := unit.Sp(20)
	// if mobile {
	// 	fontSize = unit.Sp(17.3)
	// }

	for {
		e := <-w.Events()
		giohyperlink.ListenEvents(e)

		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			for span, events := richText.Events(); span != nil; span, events = richText.Events() {
				for _, event := range events {
					content, _ := span.Content()
					switch event.Type {
					case richtext.Click:
						if event.ClickData.Type == gesture.TypeClick {
							op.InvalidateOp{}.Add(&ops)
						}
						if strings.Contains(content, "https://") {
							_ = giohyperlink.Open(content)
						} else {
							// clipboard.WriteOp{Text: content}.Add(&ops)
							// if sender, err := toast.NewSender(w); err == nil {
							// 	_ = sender.SendToast("Copied to clipboard")
							// }
						}
					}
				}
			}

			if startButton.Clicked() {
				if startText == "Start" {
					go func() {
						TARGET = address.Text() + ":" + port.Text()
						server(TARGET)
						startText = "Stop"
					}()
				}
			}

			var children []layout.FlexChild

			children = append(children,
				// The title
				layout.Rigid(func(gtx C) D {
					title := material.H1(theme, "Tedac")
					title.Color = color.NRGBA{R: 170, G: 65, B: 145, A: 255}
					title.Alignment = text.Middle
					return title.Layout(gtx)
				}),

				// The server address input
				layout.Rigid(func(gtx C) D {
					top := unit.Dp(25)
					return layout.Inset{
						Top:    top,
						Bottom: unit.Dp(25),
						Right:  unit.Dp(35),
						Left:   unit.Dp(35),
					}.Layout(gtx, func(gtx C) D {
						return material.Editor(theme, &address, "Server IP (ex: vasar.land)").Layout(gtx)
					})
				}),

				// The server port input
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    unit.Dp(5),
						Bottom: unit.Dp(5),
						Right:  unit.Dp(35),
						Left:   unit.Dp(35),
					}.Layout(gtx, func(gtx C) D {
						return material.Editor(theme, &port, "Server Port (default: 19132)").Layout(gtx)
					})
				}),

				// The start button
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(25),
						Right:  unit.Dp(35),
						Left:   unit.Dp(35),
					}.Layout(gtx, func(gtx C) D {
						return material.Button(theme, &startButton, startText).Layout(gtx)
					})
				}),

				// The debug text
				layout.Rigid(func(gtx C) D {
					return layout.Inset{
						Right: unit.Dp(35),
						Left:  unit.Dp(35),
					}.Layout(gtx, func(gtx C) D {
						return richtext.Text(&richText, theme.Shaper).Layout(gtx)
					})
				}),
			)

			macro := op.Record(gtx.Ops)
			gtx.Constraints.Min.Y = 0
			dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			call := macro.Stop()

			op.Offset(image.Point{Y: int((float32(gtx.Constraints.Max.Y)*0.95 - float32(dims.Size.Y)) / 2)}).Add(&ops)
			call.Add(&ops)

			e.Frame(gtx.Ops)
		}
	}
}
