rkt-compose: $(shell find ./cmd ./lib -name "*.go") main.go vendor
	go build

vendor: glide.lock
	glide install

glide.lock: glide.yaml
	glide update

clean:
	-rm -rf glide.lock rkt-compose vendor
