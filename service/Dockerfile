ARG ARTIFACTORY_DOCKER
FROM ${ARTIFACTORY_DOCKER}amazonlinux:2.0.20240109.0 as base

ARG BUILD_MODE=release
ENV BUILD_MODE=${BUILD_MODE}

# Builder stage
FROM base as builder

RUN yum groupinstall -y "Development Tools" && \
    yum install -y curl openssl-devel && \
    yum clean all && \
    rm -rf /var/cache/yum

RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --no-modify-path
ENV PATH="/root/.cargo/bin:${PATH}"

WORKDIR /build

COPY . .

RUN cargo build --${BUILD_MODE}

FROM base as runner

WORKDIR /app

# Copy only the built binary from the builder stage
COPY --from=builder /build/target/${BUILD_MODE}/server ./server

EXPOSE 9001

ENV LOG_LEVEL info

CMD ["./server"]
