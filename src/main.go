package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func Handler(ctx context.Context) (Response, error) {
	stream, err := http.Get("http://marco.org:8001/listen")
	//stream, err := http.Get("https://httpstat.us/401")
	if err != nil {
		log.Fatal("The request to marco.org failed")
		return Response{}, err
	}
	defer stream.Body.Close()
	
	if stream.StatusCode != http.StatusOK {
		if stream.StatusCode == http.StatusNotFound {
			log.Println("The stream is not active")
			return Response{Body: "The stream is not active", StatusCode: 200}, nil
		} 
		log.Fatal("The stream returned a unexpected code: " + strconv.Itoa(stream.StatusCode))
		return Response{Body: "The stream returned a unexpected code: " + strconv.Itoa(stream.StatusCode), StatusCode: 200}, errors.New("The stream returned a unexpected code: " + strconv.Itoa(stream.StatusCode))
	}

	log.Println("Stream is up")
	return Response{Body: "Stream is up", StatusCode: 200}, nil

}

func main() {
	lambda.Start(Handler)
}
