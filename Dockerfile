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
RUN go build -o /build/integration-testing-cli github.com/di-graph/integration-testing-cli
RUN chmod +x /build/integration-testing-cli
ENTRYPOINT ["/build/integration-testing-cli"] 