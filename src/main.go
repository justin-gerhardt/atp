package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func Handler(ctx context.Context) (Response, error) {
	if isLive() {

	}
	return Response{Body: "aaaaaaa", StatusCode: 200}, nil

}

func isLive() bool {
	stream, err := http.Get("http://marco.org:8001/listen")
	//stream, err := http.Get("https://httpstat.us/401")
	if err != nil {
		log.Fatal("The request to marco.org failed")
		return false
	}
	defer stream.Body.Close()

	if stream.StatusCode != http.StatusOK {
		if stream.StatusCode == http.StatusNotFound {
			log.Println("The stream is not active")
			return false
		}
		log.Fatal("The stream returned a unexpected code: " + strconv.Itoa(stream.StatusCode))
		return false
	}
	log.Println("The stream is live")
	return true
}

func main() {
	lambda.Start(Handler)
}
