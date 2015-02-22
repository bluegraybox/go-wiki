FROM ubuntu

RUN mkdir -p /var/local/wiki

ADD wiki /
ADD view.html /
ADD edit.html /

ENTRYPOINT /wiki

EXPOSE 80
