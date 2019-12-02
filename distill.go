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
	"github.com/jessevdk/go-flags"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	opts struct {
		BlockSize          int    `short:"b" long:"block-size" default:"50" description:"The size of the blocks the image should be distilled to. E.g. 10 will result in square blocks of 10x10 pixels. Note that this is the size of the blocks, not the number of blocks."`
		MaxDominantColours int    `short:"n" long:"max-dominant-colours" default:"8" description:"The number of dominant colours to be extracted from the image."`
		OutputPath         string `short:"o" long:"output-path" default:"out.png" description:"Path to the output file. The format will be png."`
	}
	imagePath string
)

func init() {
	args, err := flags.Parse(&opts)
	if err != nil {

		os.Exit(1)
	}
	if len(args) == 1 {
		imagePath = args[0]
	} else {
		fmt.Println("Specity an image to process")
		os.Exit(1)
	}

}

func main() {

	img, err := imgio.Open(imagePath)
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

	for x := 0; x < source.Bounds().Max.X; x += opts.BlockSize {
		for y := 0; y < source.Bounds().Max.Y; y += opts.BlockSize {
			var (
				block          = source.SubImage(image.Rect(x, y, x+opts.BlockSize, y+opts.BlockSize))
				avgBlockColour = calculateAverageBlockColour(block)
				nearestColour  = nearestColour(avgBlockColour, prominentColours)
			)
			fillBlock(x, y, opts.BlockSize, nearestColour, source)
		}
	}

	if err := imgio.Save(opts.OutputPath, source, imgio.PNGEncoder()); err != nil {
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

	colours, err := prominentcolor.KmeansWithAll(opts.MaxDominantColours, image, prominentcolor.ArgumentNoCropping, sampleWidth, nil)
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
