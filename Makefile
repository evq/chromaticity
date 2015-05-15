all: static/static.go

swagger-ui:
	git submodule init
	git submodule update

static/apidocs: swagger-ui
	cp -R swagger-ui/dist static/apidocs
	sed -i -e 's/http:\/\/petstore\.swagger\.wordnik\.com\/v2\/swagger\.json/\.\.\/apidocs\.json/' static/apidocs/index.html

static/static.go: static/apidocs
	$(eval STATIC_DIRS:=$(shell find static -type d | tr \\n ' '))
	go-bindata -pkg=static -ignore=static/static.go -o static/static.go ${STATIC_DIRS}

clean:
	rm static/static.go
