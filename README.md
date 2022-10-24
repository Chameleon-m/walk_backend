# walk_backend

# RUN
```
make api
```

# ENV
Set environment
Linux
```
# MongoDB
export MONGO_URI="mongodb://user:userpassword@localhost:27017/walk?authSource=walk&replicaSet=rs0&serverSelectionTimeoutMS=2000"
export MONGO_INITDB_ROOT_USERNAME=root
export MONGO_INITDB_ROOT_PASSWORD=rootpassword
export MONGO_INITDB_DATABASE=walk
export MONGO_INITDB_USERNAME=user
export MONGO_INITDB_PASSWORD=userpassword
export MONGO_REPLICA_SET_NAME=rs0

# COMMON
export SITE_SCHEMA=https
export SITE_HOST=localhost
export SITE_PORT=443

# GIN
export PORT=8080
export GIN_MODE=debug

# SESSION
export SESSION_SECRET=59ce2f5dc5a3f211c6f9fffb19d7cc18c098ac19645df22585c20d19477f14ae
export SESSION_NAME=session_name
export SESSION_PATH=/v1/auth
export SESSION_DOMAIN=.localhost
export SESSION_MAX_AGE=3600

# RABBITMQ
export RABBITMQ_DEFAULT_USER=guest
export RABBITMQ_DEFAULT_PASSWORD=guest
```
Or
```
cp .env.example .env
```
# DB

## MongoDB
### Gegerate keyFile
```
openssl rand -base64 700 > ./docker/mongodb/file.key
chmod 400 ./docker/mongodb/file.key
sudo chown 999:999 ./docker/mongodb/file.key
```

### Docker
Run 
```
docker-compose up -d
```
```
docker-compose down
```
Or
```
docker run --rm -d -p 27017:27017 -h $(hostname) --name mongo1 mongo:latest --replSet=rs0 && sleep 4 && docker exec mongo mongosh --eval "rs.initiate();"
```
```
docker stop mongo1
```

### Update hostnames
Once the replica set is up, you will need to update hostnames in local /etc/hosts file.
```
127.0.0.1 mongo1
```
**NOTE**: In windows, the hosts file is located at C:\Windows\System32\drivers\etc\hosts

# NGINX
Generate keys
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./.docker/nginx/keys/cert.key -out ./.docker/nginx/keys/cert.crt
```
```
openssl dhparam -out ./.docker/nginx/keys/dhparam.pem 4096
```

# RABBITMQ

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