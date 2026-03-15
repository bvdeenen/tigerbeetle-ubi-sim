tigerbeetle-ubi-sim: $(shell find . -name '*.go')
	go build -o tigerbeetle-ubi-sim .

clean:
	rm -f tigerbeetle-ubi-sim

.PHONY: clean
