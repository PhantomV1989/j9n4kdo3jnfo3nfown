# sudo docker build --network=host -t belive .
# sudo docker run --network=host belive localhost:5432 8080
FROM golang:1.14

WORKDIR /go/src/github.com/belive
COPY . /go/src/github.com/belive

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["belive"]