debug:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o debuggodin.exe -gcflags "all=-N -l" ./cmd/godin/main.go
	pwsh -ExecutionPolicy Bypass -File "$(CURDIR)/scripts/Rename-HostFile.ps1" -To Debug
	func start

deploy:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o godin.exe ./cmd/godin/main.go
	pwsh -ExecutionPolicy Bypass -File "$(CURDIR)/scripts/Rename-HostFile.ps1" -To Deploy
	func azure functionapp publish godindiscbot
