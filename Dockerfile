FROM ubuntu

ADD ./pando /opt/pando
ADD ./start.sh /opt/start.sh

RUN apt-get -qq update &&\
    apt-get -qq install -y --no-install-recommends ca-certificates curl &&\
    chmod +x /opt/pando

CMD bash /opt/start.sh
