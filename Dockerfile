FROM ubuntu

ADD wiki /
ADD view.html /
ADD edit.html /

ENTRYPOINT /wiki

EXPOSE 80
