FROM box-registry.jfrog.io/jenkins/box-centos7

LABEL com.box.name="node-alert-generator"

# Required for systemd related things to work
ENV container=docker

ADD ./build/node-alert-generator /node-alert-generator
RUN chown container:container /node-alert-generator