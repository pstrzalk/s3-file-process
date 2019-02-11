package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/oliamb/cutter"
	"image"
	"image/jpeg"
	"os"
)

const bucketName = "golang-image-process"
const regionCode = "eu-west-3"

const remoteInputPath = "remote-input.jpg"
const remoteOutputPath = "remote-output.jpg"
const localInputPath = "local-input.jpg"
const localOutputPath = "local-output.jpg"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	f, err := loadS3File(remoteInputPath, localInputPath)
	check(err)

	img, _, err := image.Decode(f)
	check(err)

	err = cropImage(img, localOutputPath)
	check(err)

	err = saveS3File(localOutputPath, remoteOutputPath)
	check(err)
}

func loadS3File(remotePath string, localPath string) (*os.File, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(regionCode)},
	)
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(sess)

	file, err := os.Create(localPath)
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(remotePath),
		})

	return file, err
}

func cropImage(img image.Image, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}

	cImg, err := cutter.Crop(img, cutter.Config{
		Height:  500,
		Width:   500,
		Mode:    cutter.TopLeft,
		Anchor:  image.Point{10, 10},
		Options: 0,
	})
	if err != nil {
		return err
	}

	opt := jpeg.Options{Quality: 80}
	err = jpeg.Encode(out, cImg, &opt)

	return err
}

func saveS3File(localPath string, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(regionCode)},
	)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(remotePath),
		Body:   file,
	})

	return err
}
