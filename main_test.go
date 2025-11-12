package main

import (
	"testing"

	"golang.org/x/sync/semaphore"
)

func TestIncreasePixelBrightness(t *testing.T) {
	val := 100
	want := int(100 * 1.25)
	pixel := Pixel{val, val, val, val}
	adjustPixelBrightness(&pixel, 1.25)

	if pixel.R != want {
		t.Errorf(`pixel red value was %d when it should be %d`, pixel.R, val)
	}
}

// pixel 0,0 should be 25% brighter (colour value should increase by 1.25)
func TestImageBrightnessIncreased(t *testing.T) {
	image := readImage("images/obaa_image.jpg")
	want := int(float32(image[0][0].R) * 1.25)
	ip := &ImageProcessor{
		pixelMap: image,
		lock:     semaphore.NewWeighted(Ulimit()),
	}
	ip.adjustImageBrightness(1.25)

	if ip.pixelMap[0][0].R != want {
		t.Errorf(`(0,0) pixel red value was %d when it should be %d`, ip.pixelMap[0][0].R, want)
	}
}
