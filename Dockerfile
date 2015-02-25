FROM ubuntu

RUN mkdir -p /var/local/wiki

ADD wiki /
ADD view.html /
ADD edit.html /
ADD config.json /

ENTRYPOINT /wiki

EXPOSE 80
