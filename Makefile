build:
	go build -ldflags "-X main.commitHash=`git rev-parse HEAD`" .
