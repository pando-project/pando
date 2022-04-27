FROM ubuntu

ADD ./pando-server /opt/pando-server
ADD start.sh /opt/start.sh

RUN apt-get clean &&\
    apt-get -qq update &&\
    apt-get -qq install -y --no-install-recommends ca-certificates curl &&\
    chmod +x /opt/pando-server

CMD bash /opt/start.sh
