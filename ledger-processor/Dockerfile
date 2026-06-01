FROM rust:1.78-alpine AS builder

RUN apk add --no-cache musl-dev pkgconfig openssl-dev

WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

COPY src ./src
RUN touch src/main.rs && cargo build --release

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/target/release/ledger-processor /ledger-processor
USER nonroot:nonroot
ENTRYPOINT ["/ledger-processor"]
