package modules

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
)

func ThumbnailImage(img *[]byte, thumbWidth uint, thumbHeight uint) (*[]byte, error) {
	srcImage, _, err := image.Decode(bytes.NewReader(*img))
	if err != nil {
		logger.Log.Panicln("Error decoding image:", err)
		return nil, err
	}
	var dstImage *image.NRGBA

	if thumbWidth == thumbHeight {
		dstImage = imaging.Thumbnail(srcImage, int(thumbWidth), int(thumbHeight), imaging.Lanczos)
	} else {
		dstImage = imaging.Fit(srcImage, int(thumbWidth), int(thumbHeight), imaging.Lanczos)
	}

	// Turn the image into
	buf := new(bytes.Buffer)
	png.Encode(buf, dstImage)
	newPngImg := buf.Bytes()

	return &newPngImg, nil
}
