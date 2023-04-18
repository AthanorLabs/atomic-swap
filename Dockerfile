FROM golang:1.19.8-alpine as builder
RUN apk add gcc musl-dev linux-headers git bash curl
RUN git clone https://github.com/AthanorLabs/atomic-swap
WORKDIR atomic-swap
RUN bash scripts/install-monero-linux.sh
RUN go build --tags=prod,netgo,osusergo --ldflags '-extldflags "-static"' ./cmd/swapd/

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /root/
COPY --from=builder /go/atomic-swap/monero-bin/monero-wallet-rpc /usr/local/bin
COPY --from=builder /go/atomic-swap/swapd .
RUN ls /usr/local/bin
RUN ./swapd --version
CMD ["/root/swapd", "--env", "stagenet", "--ethereum-endpoint", "https://rpc.sepolia.org/"]