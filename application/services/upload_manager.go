package services

import (
	"cloud.google.com/go/storage"
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

func (videoUpload *VideoUpload) UploadObject(objectPath string, client *storage.Client, ctx context.Context) error {
	log.Println("uploading file " + objectPath)
	path := strings.Split(objectPath, os.Getenv("localStoragePath")+"/")

	file, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer file.Close()

	wc := client.Bucket(videoUpload.OutputBucket).Object(path[1]).NewWriter(ctx)
	// googleapi: Error 400: Cannot insert legacy ACL for an object when uniform bucket-level access is enabled. Read more at https://cloud.google.com/storage/docs/uniform-bucket-level-access, invalid
	//wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}} // error

	if _, err := io.Copy(wc, file); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	log.Println("uploaded file " + objectPath)
	return nil
}

func (videoUpload *VideoUpload) loadPaths() error {
	err := filepath.Walk(videoUpload.VideoPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			videoUpload.Paths = append(videoUpload.Paths, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func getClientUpload() (*storage.Client, context.Context, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, nil, err
	}

	return client, ctx, nil
}

func (videoUpload *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	in := make(chan int, runtime.NumCPU())
	returnChannel := make(chan string)

	err := videoUpload.loadPaths()
	if err != nil {
		return err
	}

	uploadClient, ctx, err := getClientUpload()
	if err != nil {
		return err
	}

	for process := 0; process < concurrency; process++ {
		go videoUpload.uploadWorker(in, returnChannel, uploadClient, ctx)
	}

	go func() {
		for x := 0; x < len(videoUpload.Paths); x++ {
			in <- x
		}
		close(in)
	}()

	countDoneWorker := 0
	for r := range returnChannel {
		countDoneWorker++
		if r != "" {
			doneUpload <- r
			break
		}

		if countDoneWorker == len(videoUpload.Paths) {
			doneUpload <- "upload completed"
			break
		}
	}

	return nil
}

func (videoUpload *VideoUpload) uploadWorker(in chan int, returnChannel chan string, uploadClient *storage.Client, ctx context.Context) {
	for x := range in {
		err := videoUpload.UploadObject(videoUpload.Paths[x], uploadClient, ctx)
		if err != nil {
			videoUpload.Errors = append(videoUpload.Errors, videoUpload.Paths[x])
			log.Printf("error during the upload %v. Error: %v", videoUpload.Paths[x], err)
			returnChannel <- err.Error()
		}
		returnChannel <- ""
	}
}
