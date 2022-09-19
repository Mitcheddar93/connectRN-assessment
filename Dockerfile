# Official Docker image for golang
FROM golang as build

# Set the working directory to a docker folder in the project
WORKDIR /docker

# Copy go.mod and go.sum into the working directory, which includes the added dependencies
COPY go.mod ./
COPY go.sum ./

# Install the required Go packages for image resizing
RUN go mod download

# Copy source code into the image
COPY *.go ./
COPY files ./files

RUN CGO_ENABLED=0 go test -v

# Build the httpserver application binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /httpserver

FROM alpine
COPY --from=build /httpserver /httpserver

EXPOSE 8080

CMD ["/httpserver"]