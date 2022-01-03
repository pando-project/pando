FROM ubuntu

ADD ./pando-server /opt/pando-server
ADD start.sh /opt/start.sh
ADD ./go-swagger /opt/go-swagger
ADD ./swagger.yml /opt/swagger.yml

RUN apt-get -qq update &&\
    apt-get -qq install -y --no-install-recommends ca-certificates curl &&\
    chmod +x /opt/pando-server &&\
    chmod +x /opt/go-swagger

CMD bash /opt/start.sh
