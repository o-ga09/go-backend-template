#API用コンテナに含めるバイナリを作成するコンテナ
FROM golang:1.22-bullseye as deploy-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -trimpath -ldflags "-w -s" -o main ./service/user/main.go

#-----------------------------------------------
#API デプロイ用コンアテナ
FROM centos:centos7 as deploy-api

RUN yum -y update

EXPOSE "8081"

COPY --from=deploy-builder /app/main .

CMD ["./main"]

#-----------------------------------------------
#ローカル開発環境で利用するホットリロード環境
FROM golang:1.22 as dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go install github.com/cosmtrek/air@latest
CMD ["air"]