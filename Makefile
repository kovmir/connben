build: static

static:
	CC=musl-gcc go build --ldflags '-linkmode external -extldflags "-static"'

clean:
	rm -f ./connben

.PHONY: build static clean
