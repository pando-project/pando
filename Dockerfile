FROM alpine:3.14

ADD ./pando /opt/pando

RUN chmod +x /opt/pando && \
    mkdir /lib64 && \
    ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \

CMD /opt/pando daemon