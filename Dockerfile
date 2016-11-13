FROM fedora
MAINTAINER Kent Gustavsson <kent@minoris.se>

RUN dnf update -y
RUN dnf install nginx -y
COPY risoxy /
COPY config.toml /


EXPOSE 80 443
CMD ["/risoxy"]
