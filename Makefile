build:
	@go build -mod=readonly -ldflags="-s -w" -gcflags=all=-l -trimpath=true -o bin/goproject
