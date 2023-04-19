FROM golang:1.20 as builder

# Download monero-wallet-rpc. We need bzip2 to unpack the tar file.
RUN apt update && apt install -y bzip2
RUN arch=$(uname -m | sed 's/x86_64/linux64/; s/aarch64/linuxarm8/') && \
    curl -sSL "https://downloads.getmonero.org/cli/${arch}" -o monero.tar.bz2
RUN tar xvjf monero.tar.bz2 --no-anchored monero-wallet-rpc --strip-components=1

# Build the swapd and swapcli binaries
RUN git clone --depth=1 https://github.com/AthanorLabs/atomic-swap
WORKDIR atomic-swap
RUN make build

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /go/monero-wallet-rpc /usr/local/bin/
COPY --from=builder /go/atomic-swap/bin/ /usr/local/bin/
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid "${USER_GID}" atomic && \
    useradd --no-log-init --home-dir /atomic-swap \
    --uid "${USER_UID}" --gid "${USER_GID}" -m atomic
USER atomic
WORKDIR /atomic-swap
RUN swapd --version

# 9900 the default p2p port. swapd also listens to swapcli on 127.0.0.1:5000,
# which is not accessible outside the container by default. You have 2 options
# to interact with this RPC port:
# (1) Use swapcli inside the container::
#     $ docker exec CONTAINER_NAME_OR_ID swapcli SUBCOMMAND ...
# (2) Run the container with --network=host so 127.0.0.1:5000 is the same
#     port inside and outside of the container.
EXPOSE 9900/udp
EXPOSE 9900/tcp

CMD ["swapd", "--env", "stagenet", "--ethereum-endpoint", "https://rpc.sepolia.org/"]
