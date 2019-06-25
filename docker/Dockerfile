FROM nginx:1.17.0

RUN rm /etc/nginx/conf.d/*

RUN apt-get update \
  && apt-get install -y wget supervisor psmisc dnsutils \
  && wget -O- https://deb.nodesource.com/setup_10.x | bash - \
  && apt-get update \
  && apt-get install -y nodejs

RUN mkdir -p /opt/eremetic

#COPY ui/next.config.js /srv/www
COPY ui /srv/www

ENV NODE_ENV production

COPY eremetic /opt/eremetic/eremetic
COPY marathon.sh /marathon.sh

COPY containerfs /

ENTRYPOINT ["/usr/bin/supervisord"]
CMD [ "-c", "/etc/supervisor/supervisord.conf" ]
