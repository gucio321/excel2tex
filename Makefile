build:
	go build -ldflags "-X main.commitHash=`git describe --tag`" .
