FROM ubuntu

RUN mkdir -p /var/local/wiki

ADD wiki /
ADD view.html /
ADD edit.html /
ADD all.html /
ADD config.json /
ADD github-markdown.css /
ADD octicons.css /

ENTRYPOINT /wiki

EXPOSE 80
