FROM jessestuart/caddy-cloudflare:v0.11.0

RUN mkdir /srv/www
COPY index.html /srv/www/index.html
EXPOSE 80 443 2015
VOLUME /root/.caddy /srv
WORKDIR /srv

ENTRYPOINT ["/bin/parent", "caddy"]
CMD ["--conf", "/etc/Caddyfile", "--log", "stdout", "--agree=$ACME_AGREE"]