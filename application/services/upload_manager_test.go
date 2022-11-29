package services_test

import (
	"encoder/application/services"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func TestVideoServiceUpload(t *testing.T) {
	video, repo := prepare()
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo
	err := videoService.Download("video-encoder-microservice")
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = "video-encoder-microservice"
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(50, doneUpload)

	result := <-doneUpload
	require.Equal(t, result, "upload completed")
}
