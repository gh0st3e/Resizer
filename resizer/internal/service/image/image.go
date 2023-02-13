package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
)

const (
	Small  = "small"
	Medium = "medium"
	Large  = "large"

	PNG  = ".png"
	JPEG = ".jpeg"
	JPG  = ".jpg"

	SmallSize  = 100
	MediumSize = 300
	LargeSize  = 1000

	path = "images/"
)

type ImgService struct {
	logger *logrus.Logger
}

func NewImage(logger *logrus.Logger) *ImgService {
	return &ImgService{logger: logger}
}

func (i *ImgService) ResizeImage(file *os.File, id, dirName string) (map[string]*os.File, error) {

	ext := i.getExt(file)

	decodeImage, err := i.decodeImage(ext, file)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode image: %s", err)
	}

	sizes := i.getSizes(decodeImage)

	images, err := i.resize(sizes, ext, id, dirName)
	if err != nil {
		return nil, fmt.Errorf("couldn't resize image:%s", err)
	}

	return images, nil
}

func (i *ImgService) getExt(file *os.File) string {
	switch {
	case filepath.Ext(file.Name()) == PNG:
		i.logger.Info(".png")
		return PNG
	case filepath.Ext(file.Name()) == JPG:
		i.logger.Info(".jpg")
		return JPG
	case filepath.Ext(file.Name()) == JPEG:
		i.logger.Info(".jpeg")
		return JPG
	default:
		i.logger.Info("Unknown File Type")
		return "Unknown File Type"
	}
}

func (i *ImgService) decodeImage(ext string, file *os.File) (image.Image, error) {
	var decodeImage image.Image
	var err error

	switch ext {
	case PNG:
		decodeImage, err = i.decodePNG(file)
	case JPG:
		decodeImage, err = i.decodeJPG(file)
	default:
		return nil, fmt.Errorf("%s", ext)
	}
	if err != nil {
		return nil, err
	}

	return decodeImage, nil
}

func (i *ImgService) decodePNG(file *os.File) (image.Image, error) {
	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (i *ImgService) decodeJPG(file *os.File) (image.Image, error) {
	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (i *ImgService) getSizes(decodeImage image.Image) map[string]image.Image {
	var sizes = map[string]image.Image{
		Small:  resize.Resize(SmallSize, SmallSize, decodeImage, resize.Lanczos3),
		Medium: resize.Resize(MediumSize, MediumSize, decodeImage, resize.Lanczos3),
		Large:  resize.Resize(LargeSize, LargeSize, decodeImage, resize.Lanczos3),
	}

	return sizes
}

func (i *ImgService) resize(sizes map[string]image.Image, ext, id, dirName string) (map[string]*os.File, error) {

	var images = map[string]*os.File{
		Small:  nil,
		Medium: nil,
		Large:  nil,
	}

	for ind, v := range sizes {
		file, err := i.encode(ind, ext, id, dirName, v)
		if err != nil {
			return nil, err
		}
		images[ind] = file
	}

	return images, nil
}

func (i *ImgService) encode(ind, ext, id, dirName string, image image.Image) (file *os.File, err error) {
	out, err := os.Create(filepath.Join(dirName, ind+"_"+id))
	if err != nil {
		return nil, err
	}

	switch ext {
	case PNG:
		file, err = i.encodePNG(out, image)
	case JPG:
		file, err = i.encodeJPG(out, image)
	default:
		return nil, fmt.Errorf("%s", ext)
	}
	out.Close()

	return file, err
}

func (i *ImgService) encodePNG(file *os.File, img image.Image) (*os.File, error) {
	err := png.Encode(file, img)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (i *ImgService) encodeJPG(file *os.File, img image.Image) (*os.File, error) {
	err := jpeg.Encode(file, img, nil)
	if err != nil {
		return nil, err
	}

	return file, nil
}
