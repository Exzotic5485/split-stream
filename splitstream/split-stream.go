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
	dev, err := device.Open(ss.Device, device.WithBufferSize(1))

	if err != nil {
		log.Fatal(err)
	}

	defer dev.Close()

	if err := dev.Start(context.TODO()); err != nil {
		log.Fatal(err)
	}

	format, err := dev.GetPixFormat()

	if err != nil {
		log.Fatal(err)
	}

	if format.PixelFormat != v4l2.PixelFmtMJPEG {
		log.Fatal("invalid pixel format, requires MJPEG")
	}

	for {
		// t1 := time.Now()

		frame := <-dev.GetOutput()

		reader := bytes.NewReader(frame)

		img, err := jpeg.Decode(reader, &jpeg.DecoderOptions{})

		if err != nil {
			log.Fatal(err)
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
