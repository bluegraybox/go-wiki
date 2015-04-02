# Go Wiki

Copied, with minor revisions, from Go's [Writing Web Applications](https://golang.org/doc/articles/wiki/) page.

**WARNING:** The application itself is not secure:
* It uses a hardcoded username and password in config.json, and passes them in the clear over HTTP
* Unlike most wikis, it does not keep revision history.
* The Docker configuration files set up nginx to use HTTPS with a self-signed cert, and proxy to the wiki. In this way, you get a reasonably secure combination of username/password auth over HTTPS.

## To Do
* Copy CSS to project; cdn links seem unreliable
* Document the whole process of building and deploying this.
* provide a way to download/backup wiki pages
* provide a way to upload wiki pages in bulk
    * ssh-ing into instance might be the easiest way to do both of these and more
* Add "delete page" option
* Add "rename page" option
* Add WikiWords
* page size on mobile; larger bottoms; buttons for edit & home
* Add git commit (and push) on save? Would need cert auth to git server...

## Running in Docker
Get Docker installed and running locally ([instructions for OS X](http://docs.docker.com/installation/mac/)).
You should probably work through the tutorial to get familiar with it.

### Cross-compiling
Docker containers are (generally) 64-bit Linux systems. If you're developing on a Mac, you'll need to cross-compile the binary.

To set up Go to allow cross-compiling, you'll need to do this once:
```
cd /usr/local/go/src/
sudo GOOS=linux GOARCH=amd64 ./make.bash --no-clean
```

### Building & Running Docker
_The_ `redeployDocker` _script should do all of this for you_.

In your wiki project directory:
```
GOOS=linux GOARCH=amd64 go build
docker build -t <yourDockerHubId>/gowiki .
docker run -d -p 80:80 <yourDockerHubId>/gowiki
```
(This builds a Docker image based on the configuration in `Dockerfile`.)

NOTE: If you've built and run a gowiki image before, you'll probably want to stop and remove the old container, and delete the old image.

## Deploying to Elastic Beanstalk
**Warning: These directions are incomplete and kinda sucky.**

Elastic Beanstalk is an AWS service, so you You will need to have an AWS account set up.

More surprisingly, you will also have to have set up a Dockerhub account.
EB doesn't let you push a Docker image directly to it from your machine;
you have to specify a Dockerhub respository for it to pull from.

### Generating a self-signed certificiate
Needed for ElasticBeanstalk nginx config.
```
openssl genrsa -des3 -passout pass:x -out tmp.key 4096
openssl rsa -passin pass:x -in tmp.key -out server.key
rm tmp.key
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt
```

### Docker config for AWS Elastic Beanstalk
Elastic Beanstalk has nginx proxying to Docker, and it handles HTTPS.
More info [here](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/SSLDocker.SingleInstance.html) and [here](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/create_deploy_docker_console.html).
```
cp Dockerrun.aws.json.example Dockerrun.aws.json
```
Set the image name.

Copy `.ebextensions/ssl.config.example` to `.ebextensions/ssl.config`. Edit it and copy your cert and key blocks into it.
```
zip -r ebconfig.zip Dockerrun.aws.json .ebextensions/ssl.config
```

### Deploying Docker image to AWS Elastic Beanstalk
Once you've built and tested your Docker image locally, you can push it to Dockerhub with
```
docker push <yourDockerHubId>/gowiki
```

Create an Elastic Beanstalk instance if you haven't already. (The details of that are a little out of scope for this explanation.)

At some point, you'll be prompted to upload a configuration file. This is the `ebconfig.zip` file you just created.
The name you give it will have to be unique. It doesn't need to be a formal release number, but it will help if it's descriptive (in case you're fiddling around with config settings).

You will need to set the user and password for the wiki as Elastic Beanstalk environment variables. In the dashboard for your app, click on Configuration then the little gear next to the Software Configuration box. Add environment properties named 'username' and 'password'.
