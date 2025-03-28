run-fetch-job:
	go run ./cmd/job/main.go $(ARGS)
run-etl:
	go run ./cmd/pipeline/main.go $(ARGS)
run-server:
	go run ./cmd/server/main.go $(ARGS)
