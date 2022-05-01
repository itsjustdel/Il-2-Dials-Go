package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func uiContent(serverStarted chan bool) *fyne.Container {

	text1 := canvas.NewText("Buttons", color.White)
	startButton := widget.NewButton("Start Server", func() {
		//use channel operator to asign true
		serverStarted <- true
		fmt.Println("Server on = ", <-serverStarted)

	})

	//create the container to return
	content := container.New(layout.NewHBoxLayout(), text1, layout.NewSpacer(), startButton)

	return content
}

func main() {

	//
	patcher()

	//channels for async use
	serverStarted := make(chan bool, 1)

	//use async "goroutine" to run server. Server has a blocking call so needs to be in its own thread
	go UDPServer(serverStarted)

	myApp := app.New()
	myWindow := myApp.NewWindow("UDP Server")

	myWindow.SetContent(container.New(layout.NewVBoxLayout(), uiContent(serverStarted)))
	myWindow.ShowAndRun()

}
