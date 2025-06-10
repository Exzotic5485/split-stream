package splitstream

import (
	"bytes"
	"context"
	"image"
	"log"

	"github.com/pixiv/go-libjpeg/jpeg"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type SubImage interface {
	SubImage(r image.Rectangle) image.Image
}

type Split struct {
	Rect   image.Rectangle
	Output chan []byte
}

type SplitStream struct {
	Device string
	Splits []Split
}

func NewSplitStream(device string, splits []Split) *SplitStream {
	return &SplitStream{
		Device: device,
		Splits: splits,
	}
}

func (ss *SplitStream) Run() {
	dev, err := device.Open(ss.Device, device.WithPixFormat(v4l2.PixFormat{
		PixelFormat: v4l2.PixelFmtMJPEG,
		Width:       1920,
		Height:      1080,
	}))

	if err != nil {
		log.Fatalf("failed to initialize device %v", err)
	}

	defer dev.Close()

	if err := dev.Start(context.TODO()); err != nil {
		log.Fatalf("failed to start device streaming %v", err)
	}

	for {
		// t1 := time.Now()

		frame := <-dev.GetOutput()

		if len(frame) == 0 {
			log.Println("empty frame length")
			continue
		}

		reader := bytes.NewReader(frame)

		img, err := jpeg.Decode(reader, &jpeg.DecoderOptions{})

		if err != nil {
			log.Fatalf("failed to decode frame %v", err)
		}

		for _, split := range ss.Splits {
			var buf bytes.Buffer

			subImg := img.(SubImage).SubImage(split.Rect)

			if err := jpeg.Encode(&buf, subImg, &jpeg.EncoderOptions{
				Quality: 100,
			}); err != nil {
				log.Fatal(err)
			}

			split.Output <- buf.Bytes()
		}

		// fmt.Printf("Took %s to split frames\n", time.Since(t1))

	}
}
