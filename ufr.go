package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	version = "v0.1.0"
)

type args struct {
	Bucket string `arg:"required,-b"`
	Region string `default:"eu-west-1" arg:"-r"`
	Source string `arg:"positional,required"`
}

func (args) Version() string {
	return version
}

func closeAndCheck(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func simplifyLocation(region, uploadInfoLocation string) (string, error) {
	location, err := url.QueryUnescape(uploadInfoLocation)
	if err != nil {
		return "", err
	}
	location = strings.Replace(location, "https", "s3", 1)

	subDomain := fmt.Sprintf(".s3.%s.amazonaws.com", region)
	location = strings.Replace(location, subDomain, "", 1)
	return location, nil
}

func main() {
	var args args
	arg.MustParse(&args)

	sourceDir := args.Source
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	sourceComponents := strings.Split(sourceDir, string(os.PathSeparator))
	parents := make([]string, 0)
	for _, component := range sourceComponents {
		if len(component) > 0 {
			parents = append(parents, component)
		}
	}

	if len(sourceComponents) == 0 {
		log.Fatalf("Unable to determine immediate parent for path %s", sourceDir)
	}
	immediateParent := parents[len(parents)-1]

	clientSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(args.Region),
	}))
	uploader := s3manager.NewUploader(clientSession)

	for _, file := range files {
		fileName := file.Name()

		relPath := fmt.Sprintf("%s%c%s", sourceDir, os.PathSeparator, fileName)
		fd, err := os.Open(relPath)
		if err != nil {
			log.Fatal(err)
		}

		key := fmt.Sprintf("%s%c%s", immediateParent, os.PathSeparator, fileName)
		uploadInfo, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(args.Bucket),
			Key: aws.String(key),
			Body: fd,
		})
		if err != nil {
			log.Fatal(err)
		}
		closeAndCheck(fd)

		location, err := simplifyLocation(args.Region, uploadInfo.Location)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Copied %s as %s\n", fileName, location)
	}
}
