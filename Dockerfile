FROM golang:1.18-alpine

# copy destination
WORKDIR /app

# download dependencies 
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# copy source code
COPY . .

# build
RUN go build -o /go-batch-scheduling

# run
ENTRYPOINT [ "/go-batch-scheduling" ]

