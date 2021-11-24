FROM ubuntu

ADD ./pando /opt/pando

RUN apt-get -qq update &&\
    apt-get -qq install -y --no-install-recommends ca-certificates curl &&\
    chmod +x /opt/pando &&\
    /opt/pando init \

CMD /opt/pando daemon
