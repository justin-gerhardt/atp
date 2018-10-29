build:
#	dep ensure -v
#	env GOOS=linux go build -ldflags="-s -w" -o bin/hello src/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
#	chmod o+x ./bin/* # lambda can't run the binary unless we set the other execute permission
#	sls deploy --verbose
	img build --no-console -t atp-downloader ./downloader_task
