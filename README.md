# Go Wiki

Copied, with minor revisions, from Go's [Writing Web Applications](https://golang.org/doc/articles/wiki/) page.

**WARNING:** The application itself is not secure:
* It uses a hardcoded username and password in config.json, and passes them in the clear over HTTP
* Unlike most wikis, it does not keep revision history.
* The Docker configuration files set up nginx to use HTTPS with a self-signed cert, and proxy to the wiki. In this way, you get a reasonably secure combination of username/password auth over HTTPS.

## To Do
* Switch to HTTPS (self-signed cert)
* Add authentication (Read from wiki page - Config.txt?)
* Add git commit (and push) on save?
* Store files somewhere that can persist in AWS

## Cross-compiling
Docker containers are 64-bit Linux systems. If you're developing on a Mac, you'll need to cross-compile the binary.

To set up Go to allow cross-compiling, you'll need to do this once:
```
cd /usr/local/go/src/
sudo GOOS=linux GOARCH=amd64 ./make.bash --no-clean
```
Then in your wiki project directory:
```
GOOS=linux GOARCH=amd64 go build
```

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

Copy `.ebextensions/ssl.config.example` to `.ebextensions/ssl.config`. Edit it and copy your cert and key blocks into it.
```
zip -r ebconfig.zip Dockerrun.aws.json .ebextensions
```
Upload `ebconfig.zip` as your Elastic Beanstalk config.
