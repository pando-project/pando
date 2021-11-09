FROM ubuntu

ADD ./pando /opt/pando

RUN chmod +x /opt/pando

CMD /opt/pando daemon