FROM alpine:3.2

RUN mkdir -p /opt/eremetic

COPY eremetic /opt/eremetic/eremetic
COPY marathon.sh /marathon.sh

WORKDIR /opt/eremetic
CMD [ "/marathon.sh" ]
