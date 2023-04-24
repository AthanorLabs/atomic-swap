FROM golang:1.20 as builder

ARG VERSION=latest
RUN go install -tags=prod \
    github.com/athanorlabs/atomic-swap/cmd/bootnode@"${VERSION}" \
    github.com/athanorlabs/atomic-swap/cmd/swapcli@"${VERSION}"
RUN /go/bin/bootnode --version

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates gosu

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

# 9909 is the default p2p port. bootnode also listens to swapcli on
# 127.0.0.1:5000, which is not accessible outside the container by default. You
# have 2 options to interact with the RPC port:
# (1) Use swapcli inside the container::
#     $ docker exec CONTAINER_NAME_OR_ID swapcli SUBCOMMAND ...
# (2) Run the container with --network=host so 127.0.0.1:5000 is the same
#     port inside and outside of the container.
EXPOSE 9909/udp
EXPOSE 9909/tcp

# The swapd environment (dev, stagenet, mainnet) can be convigured via the
# SWAPD_ENV environment variable or using swapd's --env flag (which takes
# precidence). In docker, we use the environment variable to configure file
# permissions of the correct directory in the data volume. The suggested
# ways of working with this are:
# (1) Set SWAPD_ENV variable and don't use the CLI flag
# (2) Use swapd's --env=ENVIRONMENT CLI flag, but set SWAPD_ENV to the
#     identical environment or the empty string.
ENV SWAPD_ENV=stagenet
ENV SWAPD_LOG_LEVEL=info

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["bootnode"]
