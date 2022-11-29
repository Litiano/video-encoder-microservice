package services

import (
	"cloud.google.com/go/storage"
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (videoService *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(videoService.Video.FilePath)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f, err := os.Create(os.Getenv("localStoragePath") + "/" + videoService.Video.ID + ".mp4")
	if err != nil {
		return err
	}

	_, err = f.Write(body)
	if err != nil {
		return err
	}

	defer f.Close()

	log.Printf("videoService %s has been stored", videoService.Video.ID)

	return nil
}

func (videoService *VideoService) Fragment() error {
	baseDir := os.Getenv("localStoragePath") + "/"
	err := os.Mkdir(baseDir+videoService.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}

	source := baseDir + videoService.Video.ID + ".mp4"
	target := baseDir + videoService.Video.ID + ".frag"

	//cmd := exec.Command("mp4fragment", source, target)
	cmd := exec.Command("docker-compose", "exec", "-T", "app", "mp4fragment", source, target) // @todo only docker
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (videoService *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(
		cmdArgs,
		"exec", "-T", "app", "mp4dash", // @todo only docker
		os.Getenv("localStoragePath")+"/"+videoService.Video.ID+".frag",
		"--use-segment-timeline",
		"-o",
		os.Getenv("localStoragePath")+"/"+videoService.Video.ID,
		"-f",
		"--exec-dir",
		"/opt/bento4/bin/",
	)
	//cmd := exec.Command("mp4dash", cmdArgs...)
	cmd := exec.Command("docker-compose", cmdArgs...) // @todo only docker
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	printOutput(output)

	return nil
}

func (videoService *VideoService) Finish() error {
	err := os.Remove(os.Getenv("localStoragePath") + "/" + videoService.Video.ID + ".mp4")
	if err != nil {
		log.Println("error removing mp4 ", videoService.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv("localStoragePath") + "/" + videoService.Video.ID + ".frag")
	if err != nil {
		log.Println("error removing mp4 ", videoService.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(os.Getenv("localStoragePath") + "/" + videoService.Video.ID)
	if err != nil {
		log.Println("error removing video folder ", videoService.Video.ID)
		return err
	}

	log.Println("files have been removed", videoService.Video.ID)
	return nil
}

func printOutput(output []byte) {
	if len(output) > 0 {
		log.Printf("==========> Ouput: %s\n", string(output))
	}
}
