build:
	@echo "Building..."
	@cd cmd/wasm && GOOS=js GOARCH=wasm go build -o  ../../assets/db.wasm

start:
	@echo "Starting..."
	@cd cmd/server && go run main.go
