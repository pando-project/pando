FROM ubuntu

ADD ./pando /opt/pando

RUN chmod +x /opt/pando &&\
    /opt/pando init

CMD /opt/pando daemon