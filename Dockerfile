FROM scratch

ADD wiki /
ADD view.html /
ADD edit.html /
ADD rename.html /
ADD all.html /
ADD config.json /
ADD static/github-markdown.css /static/
ADD static/octicons.css /static/

ENTRYPOINT ["/wiki"]

EXPOSE 80
