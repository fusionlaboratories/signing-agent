FROM golang:alpine

RUN apk update && apk add \
        ca-certificates \
        cmake \
        g++ \
        gcc \
        git \
        make \
        libtool \
        automake \
        openssl-dev \
        linux-headers

ENV BUILD_PATH=/tmp/libs
ENV LIBRARY_PATH=$BUILD_PATH/lib
ENV C_INCLUDE_PATH=$BUILD_PATH/include
ENV LIBS_PATH=/tmp/libs
ENV LIBRARY_PATH=$LIBS_PATH/lib
ENV C_INCLUDE_PATH=$LIBS_PATH/include
ENV PROJECT_PATH=/src/github.com/qredo/dta
ENV CGO_LDFLAGS="-L $LIBS_PATH/lib64"
ENV CGO_CPPFLAGS="-I $C_INCLUDE_PATH -I $C_INCLUDE_PATH/amcl"

# build AMCL
RUN echo Building Milagro Crypt C library && \
    git clone https://github.com/apache/incubator-milagro-crypto-c.git && \
    cd incubator-milagro-crypto-c && \
    mkdir build && \
    cd build && \
    cmake \
      -D CMAKE_BUILD_TYPE=Release \
      -D BUILD_SHARED_LIBS=OFF \
      -D AMCL_CHUNK=64 \
      -D WORD_SIZE=64 \
      -D AMCL_CURVE="BLS381,SECP256K1" \
      -D AMCL_RSA="" \
      -D BUILD_PYTHON=OFF \
      -D BUILD_BLS=ON \
      -D BUILD_WCC=OFF \
      -D BUILD_MPIN=ON \
      -D BUILD_X509=OFF \
      -D CMAKE_INSTALL_PREFIX=$BUILD_PATH .. && \
    make && make install

ARG BUILD_TYPE
ARG BUILD_DATE

ENV PROJECT_PATH=/src/signing-agent

WORKDIR $PROJECT_PATH
