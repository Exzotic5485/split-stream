package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/pixiv/go-libjpeg/jpeg"
)

type SubImage interface {
	SubImage(r image.Rectangle) image.Image
}

type App struct {
	Frame       *ebiten.Image
	FrameMutex  sync.Mutex
	SendCommand chan uint8
	Focus       image.Rectangle
}

func (a *App) Update() error {
	if inpututil.IsKeyJustReleased(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		file, err := os.Create("screenshot.png")

		if err != nil {
			return err
		}

		defer file.Close()

		err = png.Encode(file, a.Frame)

		if err != nil {
			return err
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		a.Focus = image.Rect(0, 0, 1920, 1080)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		a.Focus = image.Rect(0, 0, 958, 538)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		a.Focus = image.Rect(960, 0, 1920, 538)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		a.Focus = image.Rect(0, 545, 958, 1080)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		a.Focus = image.Rect(960, 545, 1920, 1080)
	}

	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.FrameMutex.Lock()
	defer a.FrameMutex.Unlock()

	if a.Frame != nil {
		screen.DrawImage(a.Frame, nil)
	}
}

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return a.Focus.Dx(), a.Focus.Dy()
}

func main() {
	a := &App{
		SendCommand: make(chan uint8),
		Focus:       image.Rect(0, 0, 1920, 1080),
	}

	ebiten.SetWindowSize(958, 538)
	ebiten.SetWindowTitle("Split Stream")

	go handleSocket(a)

	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}

func handleSocket(app *App) {
	addr, err := net.ResolveTCPAddr("tcp", osArgOrDefault(1, "192.168.55.114:3000"))

	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	go handleOutgoing(conn, app)

	for {
		var length uint32

		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			log.Fatal(err)
		}

		frame := make([]byte, length)

		_, err := io.ReadFull(conn, frame)

		if err != nil {
			log.Fatal(err)
		}

		img, err := jpeg.Decode(bytes.NewReader(frame), &jpeg.DecoderOptions{})

		if err != nil {
			log.Fatal(err)
		}

		app.FrameMutex.Lock()
		app.Frame = ebiten.NewImageFromImage(img.(SubImage).SubImage(app.Focus))
		app.FrameMutex.Unlock()
	}
}

func handleOutgoing(conn net.Conn, app *App) {
	for {
		cmd := <-app.SendCommand

		if err := binary.Write(conn, binary.BigEndian, cmd); err != nil {
			log.Printf("failed to send command %d: %v\n", cmd, err)
		}
	}
}

func osArgOrDefault(idx int, defaultValue string) string {
	if len(os.Args)-1 < idx {
		return defaultValue
	}

	return os.Args[idx]
}
