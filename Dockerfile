FROM centos7

LABEL com.box.name="node-alert-generator"

# Required for systemd related things to work
ENV container=docker

ADD ./build/node-alert-generator /node-alert-generator
ADD config /config
RUN chown -R container:container /config  && chown container:container /node-alert-generator
