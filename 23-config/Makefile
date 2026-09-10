.PHONY: run

run:
	go run .

# same config, but env vars win over the file
run-env:
	APP_SERVER__PORT=9090 APP_LOG__LEVEL=debug go run .
