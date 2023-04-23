FROM golang:1.20 as builder

# Download monero-wallet-rpc. We need bzip2 to unpack the tar file.
RUN apt-get update && apt-get install -y bzip2
RUN arch=$(uname -m | sed 's/x86_64/linux64/; s/aarch64/linuxarm8/') && \
    curl -sSL "https://downloads.getmonero.org/cli/${arch}" -o monero.tar.bz2
RUN tar xvjf monero.tar.bz2 --no-anchored monero-wallet-rpc --strip-components=1

# Build the swapd and swapcli binaries. The BRANCH argument can be set to a
# branch, release tag, "latest", or a commit hash.
ARG VERSION=master
RUN go install -tags=prod \
    github.com/athanorlabs/atomic-swap/cmd/swapd@"${VERSION}" \
    github.com/athanorlabs/atomic-swap/cmd/swapcli@"${VERSION}"
RUN /go/bin/swapd --version

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates gosu

# /usr/local/bin has swapd, swapcli, monero-wallet-rpc and
# docker-entrypoint.sh.
COPY --from=builder /go/monero-wallet-rpc /usr/local/bin/
COPY --from=builder /go/bin/ /usr/local/bin/
COPY ./docker-entrypoint.sh /usr/local/bin/

VOLUME /data

# USER_UID and USER_GID are defined as ARGs so that, if desired, you can
# build the container with a UID equal to some user outside the container
# that will own the files in /data.
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid "${USER_GID}" atomic && \
    useradd --no-log-init --home-dir /atomic-swap \
      --uid "${USER_UID}" --gid "${USER_GID}" -m atomic && \
    ln -s /data /atomic-swap/.atomicswap

# 9900 the default p2p port. swapd also listens to swapcli on 127.0.0.1:5000,
# which is not accessible outside the container by default. You have 2 options
# to interact with this RPC port:
# (1) Use swapcli inside the container::
#     $ docker exec CONTAINER_NAME_OR_ID swapcli SUBCOMMAND ...
# (2) Run the container with --network=host so 127.0.0.1:5000 is the same
#     port inside and outside of the container.
EXPOSE 9900/udp
EXPOSE 9900/tcp

# The swapd environment (dev, stagenet, mainnet) can be convigured via the
# SWAPD_ENV environment variable or using swapd's --env flag (which takes
# precidence). In docker, we use the environment variable to configure file
# permissions of the correct directory in the data volume. The suggested
# ways of working with this are:
# (1) Set SWAPD_ENV variable and don't use the CLI flag
# (2) Use swapd's --env=ENVIRONMENT CLI flag, but set SWAPD_ENV to the
#     identical environment or the empty string.
ENV SWAPD_ENV=stagenet
ENV SWAPD_ETH_ENDPOINT=https://rpc.sepolia.org/
ENV SWAPD_LOG_LEVEL=info

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["swapd"]
