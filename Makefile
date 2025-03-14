frontend:
	cd frontend && npm run build

backend:
	GOOS=linux GOARCH=amd64 go build -o zerodrop-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o zerodrop-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o zerodrop-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o zerodrop-windows-amd64.exe

all: frontend backend
	