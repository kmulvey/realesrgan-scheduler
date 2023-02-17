REPOPATH = github.com/kmulvey/imagedup
BUILDS := cleanup client comparedirs

build: 
	for target in $(BUILDS); do \
		go build -v -ldflags="-s -w" -o ./cmd/$$target ./cmd/$$target; \
	done
