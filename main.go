package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type ImageInfo struct {
	Path            string    `json:"path"`
	Filename        string    `json:"file_name"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	Altitude        string    `json:"altitude"`
	Date            time.Time `json:"date"`
	Model           string    `json:"model"`
	PixelXDimension string    `json:"sizeX"`
	PixelYDimension string    `json:"sizeY"`
}

type FileCallback[T any] func(string) (data T, err error)

func TraverseFiles[T any](tree string, extension string, fileCallback FileCallback[T]) []T {
	var data []T
	err := filepath.Walk(tree, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fileExt := filepath.Ext(path)
			if strings.ToLower(fileExt) == extension {
				coords, err := fileCallback(path)
				if err == nil {
					data = append(data, coords)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", tree, err)
	}
	return data
}

func ReadTags(path string) (info ImageInfo, err error) {
	f, err := os.Open(path)
	filename := filepath.Base(path)
	data := ImageInfo{Path: path, Filename: filename, Lat: 0, Lng: 0}
	if err != nil {
		return data, err
	}

	x, err := exif.Decode(f)
	if err != nil {
		return data, err
	}

	camModel, err := x.Get(exif.Model)
	if err == nil {
		data.Model = camModel.String()
	}

	sizeX, err := x.Get(exif.PixelXDimension)
	if err == nil {
		data.PixelXDimension = sizeX.String()
	}

	sizeY, err := x.Get(exif.PixelYDimension)
	if err == nil {
		data.PixelYDimension = sizeY.String()
	}

	date, err := x.DateTime()
	if err != nil {
		return data, err
	}
	data.Date = date

	altitude, err := x.Get(exif.GPSAltitude)
	if err == nil {
		data.Altitude = altitude.String()
	}

	lat, lng, err := x.LatLong()
	if err != nil {
		return data, err
	}
	data.Lat = lat
	data.Lng = lng

	return data, nil
}

func main() {
	dir := os.Args[1]
	start := time.Now()
	data := TraverseFiles(dir, ".jpg", ReadTags)
	end := time.Now()
	fmt.Printf("Processed %d files in %f seconds.\n", len(data), end.Sub(start).Seconds())
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("outdata.json", b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
