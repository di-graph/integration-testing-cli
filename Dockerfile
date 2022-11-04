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

FROM golang:1.19 

# Install Packages
RUN apt-get update -q
 
WORKDIR /app

COPY . . 

RUN go mod download
# RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /integration-testing-cli github.com/di-graph/integration-testing-cli
RUN NO_DIRTY=true make build
RUN chmod +x /app/integration-testing-cli
RUN chmod +x /app/scripts/get_remote_plan_json.sh
ENTRYPOINT ["/app/integration-testing-cli"] 