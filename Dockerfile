FROM golang:1.16

WORKDIR /go/src/github.com/eallion/s3-deploy-action
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

# run in /github/workspace
CMD ["s3-deploy-action"]