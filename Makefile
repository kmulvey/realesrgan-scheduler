REPOPATH = github.com/kmulvey/imagedup
BUILDS := auto cleanup comparedirs local managecache

build: 
	for target in $(BUILDS); do \
		go build -v -ldflags="-s -w" -o ./cmd/$$target ./cmd/$$target; \
	done
