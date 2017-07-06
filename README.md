# PostgreSQL\* Service Broker for the CLOUD FOUNDRY\* Platform
---

## Usage
Using PostgreSQL requires having Go installed, also using Linux distributions
is recommended. In addition secure keypair will be required in order to deploy
this application as it works using HTTPS only(HTTP protocol not supported).

Flags:
	-key    The filepath to the ssl key.
	-cert   The filepath to the ssl certificate.

```
go get github.com/cloudfoundry-community/cf-postgresql-broker
cd $GOPATH/get github.com/cloudfoundry-community/cf-postgresql-broker
# build application
go build -ldflags "-w"
# add execution permissions
chmod u+x cf-postgresql-broker
# deploy broker by passing key and certificate files on arguments
./cf-postgresql-broker -key=keyFILE -cert=certFILE
```


## Generating key and certificate

In order to deploy the service broker using HTTPS, an x509 encoded RSA certificate will be
needed. For security, allowed certificates must be at least 2048 bit RSA, signed
with SHA256, SHA384 or SHA512 algorithms.
Using RSA:2048 with SHA384 is recommended.

To generate a RSA:2048 key pair using SHA384 use openssl command on Linux
systems.

#### RSA:2048 signed with SHA384

```
openssl req -x509 -sha384 -new -nodes -newkey rsa:2048 -keyout key.pem -out cert.pem
```

## Enable http basic Auth with Nginx\*
Once the software is running locally, you will need to enable http Basic
Authentication prior to adding the service broker on Cloud Foundry\*, you may accomplish this with Ngnix, altough there may be other
alternatives, this software is tested with Ngnix\*. The following steps assume
Ubuntu 16.04 is being used and that Nginx\* has been properly installed.

* Create additional ssl key and certificate using the instructions described at "Generating key
  and certificate", make sure to use the recommended security policies
  previously.
* Create user and password for http Basic Authentication, the tool will then
  request for a password.
```
$ sudo htpasswd -c /etc/nginx/.htpasswd someuser
```
* Create a site configuration for Nginx\*
```
sudo vi /etc/nginx/sites-available/broker.conf
```
* Enter the configuration for the broker site, you may use the following
  template and add any additional settings, remember to use the recommended
  ciphers for your key and certificate.
```
server {
	listen $PORT
	ssl_certificate $PATH_TO_CERTIFICATE;
	ssl_certificate_key $PATH_TO_CERTIFICATE_KEY;

	ssl on;

	location / {
		proxy_set_header X-Real-IP $REMOTE_ADDR;
		proxy_set_header X-Forwarded-For $REMOTE_ADDR;
		proxy_set_header Host $HOST;
		proxy_pass https://$BROKER_ADDR:$BROKER_PORT;

		auth_basic "Restricted";
		auth_basic_user_file /etc/nginx/.htpasswd;
	}
}
```
* Restart Nginx\*
```
sudo systemctl restart nginx
```


## Integration with the CLOUD FOUNDRY\* Platform
In oder to add the broker on Cloud Foundry\* you will require administrator
privileges. Follow official
[documentation](https://docs.cloudfoundry.org/services/managing-service-brokers.html) to learn how to add the broker
using "cf" tool.

## API documentation

The software follows the specification of the Service Broker API, please check
https://docs.cloudfoundry.org/services/api.html
