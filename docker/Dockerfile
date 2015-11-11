FROM alpine:3.2

RUN mkdir -p /opt/eremetic

COPY eremetic /opt/eremetic/eremetic
COPY static /opt/eremetic/static
COPY templates /opt/eremetic/templates
COPY marathon.sh /marathon.sh

WORKDIR /opt/eremetic
CMD [ "/marathon.sh" ]
