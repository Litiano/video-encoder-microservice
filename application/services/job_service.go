package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (jobService *JobService) Start() error {
	err := jobService.changeJobStatus("DOWNLOADING")
	if err != nil {
		return jobService.failJob(err)
	}
	err = jobService.VideoService.Download(os.Getenv("inputBucketName"))
	if err != nil {
		return jobService.failJob(err)
	}

	err = jobService.changeJobStatus("FRAGMENTING")
	if err != nil {
		return jobService.failJob(err)
	}
	err = jobService.VideoService.Fragment()
	if err != nil {
		return jobService.failJob(err)
	}

	err = jobService.changeJobStatus("ENCODING")
	if err != nil {
		return jobService.failJob(err)
	}
	err = jobService.VideoService.Encode()
	if err != nil {
		return jobService.failJob(err)
	}

	err = jobService.performUpload()
	if err != nil {
		return jobService.failJob(err)
	}

	err = jobService.changeJobStatus("FINISHING")
	if err != nil {
		return jobService.failJob(err)
	}
	err = jobService.VideoService.Finish()
	if err != nil {
		return jobService.failJob(err)
	}

	err = jobService.changeJobStatus("COMPLETED")
	if err != nil {
		return jobService.failJob(err)
	}

	return nil
}

func (jobService *JobService) changeJobStatus(status string) error {
	jobService.Job.Status = status
	_, err := jobService.JobRepository.Update(jobService.Job)

	if err != nil {
		return err
	}

	return nil
}

func (jobService *JobService) failJob(jobError error) error {
	jobService.Job.Status = "FAILED"
	jobService.Job.Error = jobError.Error()
	_, err := jobService.JobRepository.Update(jobService.Job)
	if err != nil {
		return err
	}

	return jobError
}

func (jobService *JobService) performUpload() error {
	err := jobService.changeJobStatus("UPLOADING")
	if err != nil {
		return jobService.failJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("outputBucketName")
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + jobService.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("concurrencyUpload"))

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(concurrency, doneUpload)

	//var uploadResult string
	uploadResult := <-doneUpload

	if uploadResult != "upload completed" {
		return jobService.failJob(errors.New(uploadResult))
	}

	return nil
}
