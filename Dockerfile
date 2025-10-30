FROM golang:tip-alpine3.22
ARG COMMIT_SHA="unknown"
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .
RUN echo "$COMMIT_SHA" > /app/version.txt
CMD ["./main"]
