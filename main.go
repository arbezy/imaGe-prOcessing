package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Pixel struct {
	R, G, B, A int
}

type ImageProcessor struct {
	pixelMap [][]Pixel
	lock     *semaphore.Weighted
}

func readImage(filepath string) [][]Pixel {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)

	file, err := os.Open("./" + filepath)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	pixels, err := getPixels(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	return pixels
}

func writeImage(img image.Image) {
	f, _ := os.Create("new_images/obaa_image.png")
	jpeg.Encode(f, img, nil)
}

func getPixels(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

func getImageFromPixels(pixels [][]Pixel) image.Image {
	width := len(pixels[0])
	height := len(pixels)

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

	for y := range height {
		for x := range width {
			c := pixelToColour(pixels[y][x])
			img.Set(x, y, c)
		}
	}

	return img
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

func pixelToColour(pixel Pixel) color.RGBA64 {
	return color.RGBA64{
		uint16(pixel.R * 257),
		uint16(pixel.G * 257),
		uint16(pixel.B * 257),
		uint16(pixel.A * 257),
	}
}

func increasePixelBrightness(pixel *Pixel, value float32) {
	pixel.R = min(255, int(float32(pixel.R)*value))
	pixel.G = min(255, int(float32(pixel.G)*value))
	pixel.B = min(255, int(float32(pixel.B)*value))
}

func (ip *ImageProcessor) increaseImageBrightness(value float32) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	fmt.Println("increasing image brightness")

	for i := 0; i < len(ip.pixelMap); i++ {
		for j := 0; j < len(ip.pixelMap[0]); j++ {
			wg.Add(1)
			ip.lock.Acquire(context.TODO(), 1)

			go func(pixel *Pixel) {
				defer ip.lock.Release(1)
				defer wg.Done()
				increasePixelBrightness(pixel, value)
			}(&ip.pixelMap[i][j])
		}
	}
}

func Ulimit() int64 {
	out, err := exec.Command("ulimit", "-n").Output()

	if err != nil {
		panic(err)
	}

	s := strings.TrimSpace(string(out))
	i, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		panic(err)
	}
	return i
}

func main() {
	image := readImage("images/obaa_image.jpg")
	ip := &ImageProcessor{
		pixelMap: image,
		lock:     semaphore.NewWeighted(Ulimit()),
	}
	ip.increaseImageBrightness(1.25)
	writeImage(getImageFromPixels(ip.pixelMap))
}
