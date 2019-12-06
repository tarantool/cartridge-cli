FROM centos:8

ARG INSTALL_PACKAGES

RUN yum -y update
RUN yum install -y yum-utils device-mapper-persistent-data lvm2
RUN yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

RUN yum install -y docker-ce-18.09.1

VOLUME /var/lib/docker

ENV PORT=2375

ADD wrapdocker /usr/local/bin/wrapdocker
RUN chmod +x /usr/local/bin/wrapdocker

RUN yum install -y $INSTALL_PACKAGES

EXPOSE 2375

ENTRYPOINT [ "/usr/local/bin/wrapdocker" ]

CMD ["/bin/bash" , "-l"]
