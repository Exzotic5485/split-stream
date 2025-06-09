package main

import (
	"bytes"
	"encoding/binary"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	Frame      *ebiten.Image
	FrameMutex sync.Mutex
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustReleased(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		file, err := os.Create("screenshot.png")

		if err != nil {
			return err
		}

		defer file.Close()

		err = png.Encode(file, g.Frame)

		if err != nil {
			return err
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		time.Sleep(time.Second)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		time.Sleep(time.Second * 2)
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		time.Sleep(time.Second * 3)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.FrameMutex.Lock()
	defer g.FrameMutex.Unlock()

	if g.Frame != nil {
		screen.DrawImage(g.Frame, nil)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Frame.Bounds().Dx(), g.Frame.Bounds().Dy()
}

func main() {
	g := &Game{}

	ebiten.SetWindowSize(958, 538)
	ebiten.SetWindowTitle("Split Stream")

	go handleSocket(g)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func handleSocket(game *Game) {
	addr, err := net.ResolveTCPAddr("tcp", osArgOrDefault(1, "192.168.55.114:3000"))

	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

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

		img, err := jpeg.Decode(bytes.NewReader(frame))

		if err != nil {
			log.Fatal(err)
		}

		game.FrameMutex.Lock()
		game.Frame = ebiten.NewImageFromImage(img)
		game.FrameMutex.Unlock()
	}
}

func osArgOrDefault(idx int, defaultValue string) string {
	if len(os.Args)-1 < idx {
		return defaultValue
	}

	return os.Args[idx]
}
