package database

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
)

func createDefaultJPEG(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, blue)
		}
	}
	return img
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeJPEG(data []byte) (image.Image, error) {
	r := bytes.NewReader(data)
	return jpeg.Decode(r)
}

func (db *appdbimpl) AddNewUser(username string, country string, city string, securityKey string) (int, error) {

	//defaultPhoto := createDefaultGIF()
	/*
		bytes, err := encodeGIF(defaultPhoto)
		if err != nil {
		return 0, err
		}
	*/
	defaultImg := createDefaultJPEG(100, 100)

	jpgBytes, err := encodeJPEG(defaultImg, 85) // quality=85
	if err != nil {
		return 0, err
	}
	res, err := db.c.Exec(`
		INSERT INTO Users (username, country, city, security_key, jpeg_photo) 
		VALUES (?, ?, ?, ?, ?)`, username, country, city, securityKey, jpgBytes)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}
