package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/eduncan911/podcast"
)

func handler(ctx context.Context, S3Event events.S3Event) error {
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "Failed to create aws session")
	}
	s3Service := s3.New(sess)
	log.Println("Handling the modification of " + strconv.Itoa(len(S3Event.Records)) + " objects")
	if len(S3Event.Records) == 0 {
		return errors.New("Event contains no records")
	}
	bucket := S3Event.Records[0].S3.Bucket.Name
	for _, record := range S3Event.Records {
		filePath := record.S3.Object.Key
		if strings.HasPrefix(record.EventName, "ObjectCreated") {
			unescapedPath, err := url.QueryUnescape(filePath)
			if err != nil {
				return errors.Wrap(err, "Error parsing file path")
			}
			err = renameEpisode(unescapedPath, s3Service, bucket)
			if err != nil {
				return errors.Wrap(err, "Failed to rename and move new episode")
			}
		}
	}
	files, err := getEpisodeFiles(s3Service, bucket)
	if err != nil {
		return errors.Wrap(err, "Error getting list of episode files")
	}
	feed := generateRSS(files)
	err = uploadRSStoS3(feed, sess, bucket)
	if err != nil {
		return errors.Wrap(err, "error generating rss feed")
	}
	return nil
}

func generateRSS(files []episodeFile) string {
	pod := podcast.New("ATP Live Broadcast", "https://atp.fm", "Episodes of ATP taken from the live broadcast", nil, nil)
	pod.AddSubTitle("Three nerds discussing tech, Apple, programming, and loosely related matters.")
	pod.AddSummary("Three nerds discussing tech, Apple, programming, and loosely related matters.")
	pod.Description = "Three nerds discussing tech, Apple, programming, and loosely related matters."
	pod.IExplicit = "no"
	pod.IAuthor = "Marco Arment, Casey Liss, John Siracusa"
	pod.IOwner = &podcast.Author{
		Name:  "atp@marco.org",
		Email: "atp@marco.org",
	}
	pod.AddCategory("Technology", nil)
	pod.AddImage("http://static1.squarespace.com/static/513abd71e4b0fe58c655c105/t/52c45a37e4b0a77a5034aa84/1388599866232/1500w/Artwork.jpg")
	pod.IBlock = "yes"
	for _, episode := range files {
		fileName := filepath.Base(episode.path)
		title := strings.TrimSuffix(fileName, ".mp3")
		pod.AddItem(podcast.Item{
			Title:       title,
			Description: title,
			PubDate:     &episode.lastModifed,
			Enclosure: &podcast.Enclosure{
				URL:    os.Getenv("BASE_URL") + "/" + episode.path,
				Length: episode.size,
				Type:   podcast.MP3,
			},
		})
	}
	//log.Println(pod.String())
	return pod.String()
}

type episodeFile struct {
	path        string
	size        int64
	lastModifed time.Time
}

func getEpisodeFiles(s3Service *s3.S3, bucket string) ([]episodeFile, error) {
	result := []episodeFile{}
	listInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: aws.String("processed/"),
	}
	err := s3Service.ListObjectsV2Pages(listInput,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, record := range page.Contents {
				result = append(result, episodeFile{
					path:        *record.Key,
					size:        *record.Size,
					lastModifed: *record.LastModified,
				})
			}
			return true
		})
	if err != nil {
		return nil, errors.Wrap(err, "Error getting list of processed files")
	}
	return result, nil
}

func uploadRSStoS3(feed string, sess *session.Session, bucket string) error {
	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &bucket,
		Key:    aws.String("feed.rss"),
		Body:   strings.NewReader(feed),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return errors.Wrap(err, "Error uploading rss feed to s3")
	}
	return nil
}

func main() {
	if os.Getenv("IS_OFFLINE") == "TRUE" {
		err := handler(nil, getExampleEvent())
		if err != nil {
			log.Printf("%s", err)
		}
		return
	}
	lambda.Start(handler)
}

func getExampleEvent() events.S3Event {
	var event events.S3Event
	data := []byte(`{"Records":[{"eventVersion":"2.0","eventSource":"aws:s3","awsRegion":"us-east-1","eventTime":"2018-11-15T04:08:43.47Z","eventName":"BLAHObjectCreated:Put","userIdentity":{"principalId":"A1EI5Z9V6UZIN8"},"requestParameters":{"sourceIPAddress":"67.193.138.73"},"responseElements":{"x-amz-id-2":"P1CuSIz75d7sjYEhT0FnmUIO82J75KtO+67/D2W+pOQAmRXmfsskJ/PvaG+N0UL70svTelMTLaw=","x-amz-request-id":"48B07CCF65CFDF38"},"s3":{"s3SchemaVersion":"1.0","configurationId":"795248a1-52d7-40bb-b813-59c1c44eee2c","bucket":{"name":"atp-episodes","ownerIdentity":{"principalId":"A1EI5Z9V6UZIN8"},"arn":"arn:aws:s3:::atp-episodes"},"object":{"key":"incoming/Text+File+%281%29.mp3","size":2,"urlDecodedKey":"","versionId":"","eTag":"d784fa8b6d98d27699781bd9a7cf19f0","sequencer":"005BECF14B6DF47EEF"}}}]}`)
	json.Unmarshal(data, &event)
	return event
}

func renameEpisode(path string, s3Service *s3.S3, bucket string) error {
	newName := getNewFileName(path)
	_, err := s3Service.CopyObject(&s3.CopyObjectInput{
		Bucket:     &bucket,
		CopySource: aws.String(bucket + "/" + url.QueryEscape(path)),
		Key:        aws.String("processed/" + newName),
		ACL:        aws.String("public-read"),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to copy episode to the processed directory")
	}
	_, err = s3Service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    aws.String(path),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to delete old episode after copying, state is inconsistent")
	}
	return nil
}

func getNewFileName(path string) string {
	fileName := filepath.Base(path)
	newEpisode, _ := regexp.MatchString("^\\d+\\.mp3$", fileName)
	if !newEpisode {
		log.Println("Created \"" + fileName + "\". This does follow new episode naming scheme, keeping current name")
		return fileName
	}
	showName, err := tryToGetShowName()
	if err != nil {
		log.Printf("%s", errors.Wrap(err, "Error when getting most recent show name, keeping current name"))
		return fileName
	}
	return showName + ".mp3"
}

func tryToGetShowName() (string, error) {
	url := url.URL{Scheme: "ws", Host: "accidentalbot.herokuapp.com", Path: "/"}
	//url := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/"}
	wsConnection, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "Error establishing connection with "+url.String())
	}
	defer wsConnection.Close()
	var message struct {
		Operation string
		Titles    []struct {
			ID     int
			Author string
			Title  string
			Votes  int
			Voted  bool
			Time   string
		}
		Links []struct {
			ID     int
			Author string
			Link   string
			Time   string
		}
	}
	err = wsConnection.ReadJSON(&message)
	if err != nil {
		return "", errors.Wrap(err, "Error reading showbot message")
	}
	err = wsConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return "", errors.Wrap(err, "Error closing connection to showbot")
	}
	if message.Operation != "REFRESH" {
		return "", errors.New("Showbot message was not a refresh")
	}
	if len(message.Titles) == 0 {
		return "", errors.New("Showbot had no titles")
	}
	sort.Slice(message.Titles, func(i, j int) bool { return message.Titles[i].Votes > message.Titles[j].Votes })
	mostVoted := message.Titles[0]
	log.Println("Most voted title is \"" + mostVoted.Title + "\"")
	return mostVoted.Title, nil
}
