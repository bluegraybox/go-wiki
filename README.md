# Go Wiki

Copied, with minor revisions, from Go's [Writing Web Applications](https://golang.org/doc/articles/wiki/) page.

**WARNING:** This is *totally* insecure: no HTTPS, no authentication, no revision history, nothing.

## To Do
* Switch to HTTPS (self-signed cert)
* Add authentication (Read from wiki page - Config.txt?)
* Add git commit (and push) on save?
* Store files somewhere that can persist in AWS

## Running in Docker
```
docker build -t <yourDockerHubId>/gowiki .
docker run -d -p 80:80 <yourDockerHubId>/gowiki
```

## Generating a self-signed certificiate
```
openssl genrsa -des3 -passout pass:x -out tmp.key 4096
openssl rsa -passin pass:x -in tmp.key -out server.key
rm tmp.key
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt
```

## Docker config for AWS Elastic Beanstalk
Elastic Beanstalk has nginx proxying to Docker, and it handles HTTPS.
```
cp Dockerrun.aws.json.example Dockerrun.aws.json
```
Set the image name.
```
zip -r ebconfig.zip Dockerrun.aws.json .ebextensions
```
Upload `ebconfig.zip` as your Elastic Beanstalk config.
