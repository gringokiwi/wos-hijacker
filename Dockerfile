FROM openresty/openresty:alpine

RUN apk add --no-cache git \
    && /usr/local/openresty/luajit/bin/luarocks install lua-resty-http \
    && mkdir -p /usr/local/openresty/nginx/conf/conf.d

COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY default.conf /usr/local/openresty/nginx/conf/conf.d/default.conf
COPY fetch.lua /usr/local/openresty/nginx/conf/fetch.lua

EXPOSE 8080

CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]