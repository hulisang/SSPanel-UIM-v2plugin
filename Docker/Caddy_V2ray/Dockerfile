FROM caddy:latest

RUN mkdir /srv/www && mkdir /srv/www/js
COPY index.html /srv/www/index.html
ADD js     /srv/www/js
COPY fly.css /src/www/fly.css
EXPOSE 80 443 2015
VOLUME /root/.caddy /srv
WORKDIR /srv

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
