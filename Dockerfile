# Start from a Debian image with the latest version of Go 1.7 installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.7

# Set carshare-back to look for mongo server at "mongo" url for easier linking
ENV CARSHARE_MGO_URL mongo

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/LewisWatson/carshare-back

# Build the carshare-back command inside the container.
RUN go get github.com/benbjohnson/clock
RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/manyminds/api2go
RUN go get gopkg.in/mgo.v2
RUN go install github.com/LewisWatson/carshare-back

# Run the carshare-back command by default when the container starts.
ENTRYPOINT /go/bin/carshare-back

# Document that the service listens on port 31415.
EXPOSE 31415
