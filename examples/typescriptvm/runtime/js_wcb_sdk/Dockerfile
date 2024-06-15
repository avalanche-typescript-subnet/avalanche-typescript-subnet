FROM rust

WORKDIR /usr/src
RUN apt update
RUN DEBIAN_FRONTEND=noninteractive apt install curl pkg-config libssl-dev clang -y

RUN git clone https://github.com/crrrazyden/javy.git

WORKDIR /usr/src/javy
RUN cargo build -p javy-core --target=wasm32-wasi -r
RUN cargo build -p javy-cli -r

RUN mkdir /out
VOLUME [ "/out" ]

ENTRYPOINT [ "/usr/src/javy/target/release/javy" ] 