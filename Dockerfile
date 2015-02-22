FROM ubuntu

RUN mkdir -p /var/local/wiki

ADD wiki /
ADD view.html /
ADD edit.html /
ADD server.crt /
ADD server.key /

ENTRYPOINT /wiki

EXPOSE 443
