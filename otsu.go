package gotsu

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
)

func Binarize(r io.Reader, w io.Writer) error {
	// decode image
	img, _, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// create gray scale image
	gray := image.NewGray(img.Bounds())
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grayColor := color.Gray16{uint16(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))}
			gray.Set(x, y, grayColor)
		}
	}

	// apply otsu binarization
	threshold := getOtsuThreshold(gray)
	binary := image.NewGray(gray.Bounds())
	for x := 0; x < gray.Bounds().Max.X; x++ {
		for y := 0; y < gray.Bounds().Max.Y; y++ {
			if gray.GrayAt(x, y).Y > threshold {
				binary.Set(x, y, color.Gray{255})
			} else {
				binary.Set(x, y, color.Gray{0})
			}
		}
	}

	if err := jpeg.Encode(w, binary, nil); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}
	return nil
}

// getOtsuThreshold returns the threshold value for binarization
func getOtsuThreshold(img *image.Gray) uint8 {
	histogram := make([]int, 256)
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			histogram[img.GrayAt(x, y).Y]++
		}
	}

	sum := 0
	for i := 0; i < 256; i++ {
		sum += i * histogram[i]
	}

	sumB := 0
	wB := 0
	wF := 0
	var max float64
	var threshold uint8

	total := img.Bounds().Dx() * img.Bounds().Dy()

	for i := 0; i < 256; i++ {
		wB += histogram[i]
		if wB == 0 {
			continue
		}
		wF = total - wB
		if wF == 0 {
			break
		}

		sumB += i * histogram[i]

		mB := float64(sumB) / float64(wB)
		mF := float64(sum-sumB) / float64(wF)

		between := float64(wB) * float64(wF) * math.Pow(mB-mF, 2)
		if between > max {
			max = between
			threshold = uint8(i)
		}
	}
	return threshold
}
