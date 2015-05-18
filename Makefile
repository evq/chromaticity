all: static/static.go

xc:
	docker run -it -v $(shell pwd):/build -v ${GOPATH}:/gopath slim-wink-buildroot
	mv build chromaticity

swagger-ui:
	git submodule init
	git submodule update

static/apidocs: swagger-ui
	cp -R swagger-ui/dist static/apidocs
	rm static/apidocs/index.html
	$(MAKE) patch-static

.PHONY: patch-static
patch-static:
	cp static/patch/index.html static/apidocs/index.html
	cp static/patch/arrive.min.js static/apidocs/lib/arrive.min.js
	cp static/patch/screen.css static/apidocs/css/screen.css
	cp static/patch/logo_small.png static/apidocs/images/logo_small.png

static/static.go: static/apidocs
	$(eval STATIC_DIRS:=$(shell find static -type d | tr \\n ' '))
	go-bindata -pkg=static -ignore=static/static.go -o static/static.go ${STATIC_DIRS}

clean:
	rm static/static.go
