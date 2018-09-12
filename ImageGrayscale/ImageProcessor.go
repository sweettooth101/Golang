package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

const debug = true

//----------------- Utilities ---------------------------//
func put(params ...interface{}) {
	if debug {
		fmt.Println(params...)
	}
}

func putf(format string, values ...interface{}) {
	if debug {
		fmt.Println(fmt.Sprintf(format, values...))
	}
}

func parseInput(inputPath string, prefix string) interface{} {

	r, err := regexp.Compile("([^\\\\/]+)\\.(jpeg|jpg)$")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	matches := r.FindStringSubmatch(inputPath)
	if len(matches) != 3 {
		//log.Fatal("invalid inputPath path: " + inputPath)
		putf("invalid inputPath path: %q ", inputPath)
		return nil
	}

	outputPath, err := filepath.Abs("./images/" + matches[1] + "_gray_" + prefix + "." + matches[2])
	if err != nil {
		log.Fatal("outputPath error!")
		return nil
	}

	return outputPath
}

//----------------- No Goroutines ---------------------------//

//GrayScaleExecute is...
func GrayScaleExecute(inputPath string, outputPath string) error {

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputImage, err := jpeg.Decode(inputFile)
	if err != nil {
		return err
	}

	rect := inputImage.Bounds()
	outputImage := image.NewRGBA(rect)
	for y := 0; y < rect.Max.Y; y++ {
		for x := 0; x < rect.Max.X; x++ {
			outputImage.Set(x, y, color.GrayModel.Convert(inputImage.At(x, y)))
		}
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return jpeg.Encode(outputFile, outputImage, nil)

}

//GrayScale is...
func GrayScale(inputPath string) error {
	if res := parseInput(inputPath, "plain"); res != nil {
		outputPath, ok := res.(string)
		if !ok {
			return errors.New("parseInput: an error occurred #2")
		}

		putf("SRC: %q  \n--->\nDES: %q ", inputPath, outputPath)
		return GrayScaleExecute(inputPath, outputPath)
	}

	return errors.New("parseInput: an error occurred #1")
}

//-------------------- Goroutines ---------------------//

//---- render row by row

//GrayScaleConcurrencyExcuteRbr is...
func GrayScaleConcurrencyExcuteRbr(inputPath string, outputPath string, numOfWoker int) error {

	//read input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputImage, err := jpeg.Decode(inputFile)
	if err != nil {
		return err
	}

	//render grayscale image
	rect := inputImage.Bounds()
	outputImage := image.NewRGBA(rect)
	maxY := rect.Max.Y
	//maxY := 6
	maxX := rect.Max.X

	renderOneRow := func(y int) {
		for x := 0; x < maxX; x++ {
			r, g, b, _ := inputImage.At(x, y).RGBA()
			factor := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
			(*outputImage).Set(x, y, color.Gray{factor})
		}
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(maxY)

	channelRows := make(chan int, maxY)
	for i := 0; i < maxY; i++ {
		channelRows <- i
	}

	close(channelRows)

	var renderFunc func()
	renderFunc = func() {
		go func() {
			select {
			case y, ok := <-channelRows:
				if ok {
					renderOneRow(y)
					//putf("done %v", y)
					waitGroup.Done()

					renderFunc()

				} else {
					//put("channelRows: Channel closed!")
				}
			default:
				//put("No value ready.")
			}

		}()
	}

	//for i := 0; i < 10; i++ {
	for i := 0; i < numOfWoker && i < maxY; i++ {
		renderFunc()
	}

	//put("------- wait ---------")
	waitGroup.Wait()
	//put("------- all done -------")

	//write to output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return jpeg.Encode(outputFile, outputImage, nil)

}

//GrayScaleConcurrencyRbr is
func GrayScaleConcurrencyRbr(inputPath string, numOfWoker int) error {

	if res := parseInput(inputPath, "rbr"); res != nil {
		outputPath, ok := res.(string)
		if !ok {
			return errors.New("parseInput: an error occurred #2")
		}
		putf("SRC: %q  \n--->\nDES: %q , Wokers: %v  ", inputPath, outputPath, numOfWoker)
		return GrayScaleConcurrencyExcuteRbr(inputPath, outputPath, numOfWoker)
	}

	return errors.New("parseInput: an error occurred #1")
}

//---render parts

//GrayScaleConcurrencyExcuteParts is...
func GrayScaleConcurrencyExcuteParts(inputPath string, outputPath string, numOfWoker int) error {

	//read input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputImage, err := jpeg.Decode(inputFile)
	if err != nil {
		return err
	}

	//render grayscale image
	rect := inputImage.Bounds()
	outputImage := image.NewRGBA(rect)
	maxY := rect.Max.Y
	//maxY := 100
	//maxY := 6
	maxX := rect.Max.X

	div := int(maxY / numOfWoker)
	rest := (maxY % numOfWoker)
	restFrom := numOfWoker * div

	//putf("maxY %v , div %v , rest %v", maxY, div, rest)

	renderRows := func(fromY int, toY int) {
		for y := fromY; y <= toY; y++ {
			for x := 0; x < maxX; x++ {
				r, g, b, _ := inputImage.At(x, y).RGBA()
				factor := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
				(*outputImage).Set(x, y, color.Gray{factor})
			}
		}
	}

	var waitGroup sync.WaitGroup
	if div > 0 {

		waitGroup.Add(numOfWoker)

		var renderFunc func(from int, to int)
		renderFunc = func(from int, to int) {
			go func() {
				renderRows(from, to)
				//putf("done %v,%v", from, to)
				waitGroup.Done()

			}()
		}

		for i := 0; i < numOfWoker; i++ {
			from := i * div
			to := from + div - 1
			//putf("rendering part: [%v,%v]", from, to)
			renderFunc(from, to)
		}

	}

	if rest > 0 {

		from := restFrom
		to := restFrom + rest - 1

		//putf("done rest: [%v,%v]", from, to)
		renderRows(from, to)
	}

	if div > 0 {
		//put("------- wait group ---------")
		waitGroup.Wait()
		//put("------- all done -------")
	}

	//write to output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return jpeg.Encode(outputFile, outputImage, nil)

}

//GrayScaleConcurrencyParts is
func GrayScaleConcurrencyParts(inputPath string, numOfWoker int) error {

	if res := parseInput(inputPath, "pt"); res != nil {
		outputPath, ok := res.(string)
		if !ok {
			return errors.New("parseInput: an error occurred #2")
		}
		putf("SRC: %q  \n--->\nDES: %q , Wokers: %v  ", inputPath, outputPath, numOfWoker)
		return GrayScaleConcurrencyExcuteParts(inputPath, outputPath, numOfWoker)
	}

	return errors.New("parseInput: an error occurred #1")
}

func main() {

	path := flag.String("path", "", "Path to input image file")
	maxgr := flag.Int("maxgr", 0, "Maximum nunber of Gorountines")
	al := flag.Int("al", 0, "Which algorithm to use for ")

	flag.Parse()

	if *maxgr == 0 {
		GrayScale(*path)
	} else {
		if *maxgr > 0 {
			if *al == 0 {
				GrayScaleConcurrencyRbr(*path, *maxgr)

			} else {
				GrayScaleConcurrencyParts(*path, *maxgr)
			}

		}
	}

	//GrayScaleConcurrencyExcuteRbr("./images/1.jpg", "./images/ben_rbr5_grayout.jpg", 5)
	//GrayScaleConcurrencyExcuteParts("./images/1.jpg", "./images/ben_pt5_grayout.jpg", 5)

}
