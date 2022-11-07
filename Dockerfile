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

# Set Environment Variables
SHELL ["/bin/bash", "-c"]
ENV HOME /app
ENV CGO_ENABLED 0

# Install Packages
RUN apt-get update -q
RUN apt-get install bash
 
WORKDIR /app

COPY . .
RUN go mod download
RUN NO_DIRTY=true make release
RUN NO_DIRTY=true make build
RUN chmod +x /app/integration-testing-cli
ENTRYPOINT ["/app/integration-testing-cli"] 