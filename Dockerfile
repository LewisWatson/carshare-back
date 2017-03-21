# Start from a Debian image with the latest version of Go 1.8 installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8

# Set carshare-back to look for mongo server at "mongo" url for easier linking
ENV CARSHARE_MGO_URL mongo

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/LewisWatson/carshare-back

# Download and install any required third party dependencies into the container.
RUN go get ./...

# Build the carshare-back command inside the container.
RUN go install github.com/LewisWatson/carshare-back

# Run the carshare-back command by default when the container starts.
ENTRYPOINT /go/bin/carshare-back

# Document that the service listens on port 31415.
EXPOSE 31415
