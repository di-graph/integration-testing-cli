# FROM golang:1.19 as builder

# WORKDIR /src

# # Build Application
# COPY . .
# RUN make deps
# RUN make build
# RUN chmod +x /src/build/integration-testing-cli

# # Application
# FROM alpine:3.16 as app

# ENTRYPOINT [ "integration-testing-cli" ]

FROM golang:alpine 
 
WORKDIR /app

COPY . . 

RUN go mod download
# RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /integration-testing-cli github.com/di-graph/integration-testing-cli
RUN NO_DIRTY=true make build
RUN chmod +x /integration-testing-cli
ENTRYPOINT ["/integration-testing-cli"] 