FROM golang

# it is okay to leave user/GoDoRP as long as you do not want to share code with other libraries
COPY . /go/src/bt_txbyaddr
WORKDIR /go/src/bt_txbyaddr

# added vendor services will need to be included here

RUN go get github.com/lib/pq

RUN go get ./
RUN go build
CMD ["bt_txbyaddr"]
EXPOSE 8080