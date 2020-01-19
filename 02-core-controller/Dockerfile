FROM golang:1.12.15 AS builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
WORKDIR /app

# Get dependancies - will be cached if we won't change mod/sum
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . /app/
RUN make clean && make

FROM centos:7
COPY --from=builder /app/job /job
ENTRYPOINT [ "/job" ]