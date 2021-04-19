.PHONY: build

build-NotifyEventHandlerFn:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(ARTIFACTS_DIR)/handler src/notify-event-handler/main.go
build-NotifyGraphqlApiFn:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(ARTIFACTS_DIR)/handler src/notify-graphql-api/main.go
build-UpdateStateStoreFn:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(ARTIFACTS_DIR)/handler src/update-state-store/main.go
