APP?=app
ImageName?=justgps/aiserver
version?=0.1.0
ContainerName?=jobs
PORT?=11010
MKFILE := $(abspath $(lastword $(MAKEFILE_LIST)))
CURDIR := $(dir $(MKFILE))
GoMode?=off

clean:
	go clean
tidy:
	go mod tidy

build: tidy clean
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=${GoMode} go build -tags netgo \
	-o ${APP}

docker: build
	docker build -t ${ImageName}:${version} .
	rm -f ${APP}
	docker images
	docker push ${ImageName}:${version}

log:
	docker logs -f -t --tail 20 ${ContainerName}

win:
	GOOS=windows GOARCH=amd64 go build -o app.exe .

run: docker
	docker run -d --restart=always --name ${ContainerName} \
	-v /etc/localtime:/etc/localtime:ro \
	-v /etc/ssl/certs:/etc/ssl/certs \
	-v /etc/pki/ca-trust/extracted/pem:/etc/pki/ca-trust/extracted/pem \
	-v /etc/pki/ca-trust/extracted/openssl:/etc/pki/ca-trust/extracted/openssl \
	-v ${CURDIR}www:/app/www  \
	-v ${CURDIR}envfile:/app/envfile  \
	-v ${CURDIR}tmp:/app/tmp  \
	-p ${PORT}:80 \
	--env-file ${CURDIR}envfile \
	${ImageName}:${version}
	sh clean.sh
	clear
	make log

stop:
	docker stop ${ContainerName}
	docker rm ${ContainerName}

re: stop run

s:
	git push -u origin main
