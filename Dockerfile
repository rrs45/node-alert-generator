FROM box-registry.jfrog.io/jenkins/box-centos7

LABEL com.box.name="alert-generator"

# Required for systemd related things to work
ENV container=docker

ADD ./build/alert-generator /alert-generator
RUN chown container:container /alert-generator