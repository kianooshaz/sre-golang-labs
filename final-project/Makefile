.PHONY: build vet test run import stats demo-data clean

build:
	go build -o bin/logan ./cmd/logan

vet:
	go vet ./...

test:
	go test -race ./...

run:
	go run ./cmd/logan serve

import-sample:
	go run ./cmd/logan import testdata/sample.log

stats:
	go run ./cmd/logan stats

# A >=10k-line sample for the graded demo (README -> "The demo").
demo-data:
	awk 'BEGIN{
		for (i = 0; i < 12000; i++) {
			svc = "billing"; lvl = "INFO"; msg = "invoice processed";
			m = i % 10;
			if (m == 3 || m == 7) { lvl = "ERROR"; msg = "payment failed for order " i }
			if (m == 5)           { lvl = "WARN";  msg = "slow query " (i * 37) "ms" }
			if (m == 9)           { svc = "checkout"; lvl = "ERROR"; msg = "timeout connecting to 10.0.0." (i % 250) }
			printf "2026-09-10T09:%02d:%02dZ %s [%s] %s\n", i % 60, i % 60, lvl, svc, msg
		}
	}' > testdata/big.log

clean:
	rm -rf bin logan.db testdata/big.log
