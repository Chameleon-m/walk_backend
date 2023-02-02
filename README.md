# walk_backend

# MAIN COMMANDS
```
make build
make test
make lints
make migrate-up
make swagger-generate
make build-mocks
# ... see Makefile
```

# ENV
See .env.example

Linux add env example, add to ~/.profile
```
# COMMON
export SITE_SCHEMA=https
export SITE_HOST=localhost
export SITE_PORT=443
# ... from .env.example
```

Windows add env example via GUI, System->Advanced system params->Environment Variables
# DB

## MongoDB
### Gegerate keyFile
Linux
```
cd deployments/docker/mongodb
cp file.example.key file.key
openssl rand -base64 700 > file.key
# if permission problem
chmod 400 file.key
sudo chown 999:999 file.key
```
Windows
``` 
cd deployments/docker/mongodb
cp file.example.key file.key
openssl rand -base64 700 > file.key
# if permission problem
icacls.exe file.key /reset
icacls.exe file.key /grant:r "$($env:username):(r)"
icacls.exe file.key /inheritance:r
```

### Docker
Run 
```
docker-compose up -d
```
```
docker-compose down
```

# NGINX
Generate keys
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./.docker/nginx/keys/cert.key -out ./.docker/nginx/keys/cert.crt
```
```
openssl dhparam -out ./.docker/nginx/keys/dhparam.pem 4096
```
Or use example keys

# SWAGGER

```
make swagger-generate
```
```
make swagger-serve
```
```
make swagger-serve-f
```

### Update hostnames
Once the replica set is up, you will need to update hostnames in local /etc/hosts file.
```
127.0.0.1 grafana prometheus
# ...
```
**NOTE**: In windows, the hosts file is located at C:\Windows\System32\drivers\etc\hosts