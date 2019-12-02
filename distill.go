package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/anthonynsimon/bild/clone"
	"github.com/anthonynsimon/bild/imgio"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	blockSize          = 50
	maxDominantColours = 8
)

func main() {
	img, err := imgio.Open("starry-night.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	source := clone.AsRGBA(img)

	prominentColours, err := extractProminentColours(source)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for x := 0; x < source.Bounds().Max.X; x += blockSize {
		for y := 0; y < source.Bounds().Max.Y; y += blockSize {
			var (
				block          = source.SubImage(image.Rect(x, y, x+blockSize, y+blockSize))
				avgBlockColour = calculateAverageBlockColour(block)
				nearestColour  = nearestColour(avgBlockColour, prominentColours)
			)
			fillBlock(x, y, blockSize, nearestColour, source)
		}
	}

	if err := imgio.Save("out.png", source, imgio.PNGEncoder()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fillBlock(blockX, blockY, blockSize int, fillColour color.Color, image *image.RGBA) {
	for x := blockX; x <= blockX+blockSize; x++ {
		for y := blockY; y <= blockY+blockSize; y++ {
			image.Set(x, y, fillColour)
		}
	}
}

func extractProminentColours(image image.Image) ([]colorful.Color, error) {
	var prominentcolors []colorful.Color
	sampleWidth := uint(image.Bounds().Max.X / 2)

	colours, err := prominentcolor.KmeansWithAll(maxDominantColours, image, prominentcolor.ArgumentNoCropping, sampleWidth, nil)
	if err != nil {
		return nil, err
	}

	// convert prominentcolor.ColourItem to colourful.Color
	for _, colourItem := range colours {
		golangColor := color.RGBA{
			uint8(colourItem.Color.R),
			uint8(colourItem.Color.G),
			uint8(colourItem.Color.B),
			255}

		colourfulColor, _ := colorful.MakeColor(golangColor)
		prominentcolors = append(prominentcolors, colourfulColor)
	}

	return prominentcolors, nil
}

func nearestColour(colour color.Color, possibleColours []colorful.Color) color.Color {
	var (
		nearestColour   color.Color
		closestDistance = math.MaxFloat64
	)

	sourceColour, _ := colorful.MakeColor(colour)

	for _, possibleColour := range possibleColours {
		var distance = sourceColour.DistanceCIEDE2000(possibleColour)
		if distance < closestDistance {
			closestDistance = distance
			nearestColour = possibleColour
		}
	}
	return nearestColour
}

func calculateAverageBlockColour(block image.Image) color.Color {
	var avgR, avgG, avgB uint32

	bounds := block.Bounds()

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := block.At(x, y).RGBA()
			avgR += r
			avgG += g
			avgB += b
		}
	}

	totalPixels := uint32(bounds.Dy() * bounds.Dx())
	avgR /= totalPixels
	avgG /= totalPixels
	avgB /= totalPixels

	return color.RGBA{uint8(avgR / 0x101), uint8(avgG / 0x101), uint8(avgB / 0x101), 255}
}
