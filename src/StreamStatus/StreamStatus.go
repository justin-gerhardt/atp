package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/pkg/errors"
)

func isDownloaderTaskAlreadyRunning(ecsService *ecs.ECS) (bool, error) {
	var nextToken *string
	for {
		list, err := ecsService.ListTasks(&ecs.ListTasksInput{
			Cluster:       aws.String(os.Getenv("CLUSTER")),
			DesiredStatus: aws.String("RUNNING"),
			NextToken:     nextToken,
		})
		if err != nil {
			return false, errors.Wrap(err, "Failed to list tasks on atp cluster")
		}
		nextToken = list.NextToken
		if len(list.TaskArns) == 0 {
			break
		}
		tasks, err := ecsService.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(os.Getenv("CLUSTER")),
			Tasks:   list.TaskArns,
		})
		if err != nil {
			return false, errors.Wrap(err, "Failed to describe running tasks")
		}
		for _, task := range tasks.Tasks {
			if *task.TaskDefinitionArn == os.Getenv("TASK_DEFINITION") {
				log.Println("Downloader task is already running")
				return true, nil
			}
		}
		if nextToken == nil {
			break
		}
	}
	log.Println("Downloader task not already running")
	return false, nil
}

func launchDownloaderTask() error {
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "Failed to create aws session")
	}
	ecsService := ecs.New(sess)
	isAlreadyRunning, err := isDownloaderTaskAlreadyRunning(ecsService)
	if err != nil {
		return errors.Wrap(err, "Failed to check if download task is already running")
	}
	if !isAlreadyRunning {
		_, err := ecsService.RunTask(&ecs.RunTaskInput{
			Cluster:    aws.String(os.Getenv("CLUSTER")),
			Count:      aws.Int64(1),
			LaunchType: aws.String(ecs.LaunchTypeFargate),
			NetworkConfiguration: &ecs.NetworkConfiguration{
				AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
					AssignPublicIp: aws.String(ecs.AssignPublicIpEnabled),
					Subnets:        []*string{aws.String(os.Getenv("SUBNET"))},
					SecurityGroups: []*string{aws.String(os.Getenv("SECURITY_GROUP"))},
				},
			},
			TaskDefinition: aws.String(os.Getenv("TASK_DEFINITION")),
		})
		if err != nil {
			return errors.Wrap(err, "Failed to start task")
		}
	}
	return nil
}

func handler(ctx context.Context) error {
	live, err := isLive()
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	if live {
		err := launchDownloaderTask()
		if err != nil {
			log.Printf("%s", err)
			return err
		}
	}
	return nil

}

func isLive() (bool, error) {
	stream, err := http.Get("https://atp.fm:8443/listen")
	//stream, err := http.Get("https://httpstat.us/401")
	if err != nil {
		return false, errors.Wrap(err, "Can't check if stream is online. The request failed")
	}
	defer stream.Body.Close()
	switch stream.StatusCode {
	case http.StatusNotFound:
		log.Println("The stream is not active")
		return false, nil
	case http.StatusOK:
		log.Println("The stream is live")
		return true, nil
	default:
		return false, errors.New("The stream returned a unexpected code: " + strconv.Itoa(stream.StatusCode))
	}
}

func main() {
	if os.Getenv("IS_OFFLINE") == "TRUE" {
		handler(nil)
		return
	}
	lambda.Start(handler)
}
