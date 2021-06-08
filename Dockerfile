FROM ubuntu:18.04
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
RUN mkdir /document-cache-code
WORKDIR /document-cache-code
RUN apt update
RUN apt install apt-utils curl -y
RUN curl -O https://dl.google.com/go/go1.15.8.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.15.8.linux-amd64.tar.gz
COPY . /document-cache-code
RUN tar -C /usr/local -xzf go1.15.8.linux-amd64.tar.gz
RUN /usr/local/go/bin/go install
