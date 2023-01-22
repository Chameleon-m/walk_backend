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

# LOG trace -1 debug 0 info 1 warn 2 error 3  fatal 4 panic 5 disabled = 6 "" = 7
export LOG_DEFAULT_LEVEL=debug
export LOG_CLIENT_LEVEL=warn
export LOG_SERVER_LEVEL=error

# API # HOST host.docker.internal or api or ip
export API_VERSION=0.0.1
export API_SCHEMA=http
export API_HOST=api
export API_PORT=8080

# GIN
export PORT=8080
export GIN_MODE=debug

# SESSION
export SESSION_SECRET=59ce2f5dc5a3f211c6f9fffb19d7cc18c098ac19645df22585c20d19477f14ae
export SESSION_NAME=session_name
export SESSION_PATH=/api/v1/auth
export SESSION_DOMAIN=.localhost
export SESSION_MAX_AGE=3600

# RABBITMQ
export RABBITMQ_URI="amqp://guest:guset@localhost:5672/"
export RABBITMQ_DEFAULT_USER=guest
export RABBITMQ_DEFAULT_PASSWORD=guest
export RABBITMQ_EXCHANGE_REINDEX=reindex_exchange
export RABBITMQ_ROUTING_PLACE_KEY=place_routing_key
export RABBITMQ_QUEUE_PLACE_REINDEX=place_reindex_queue

# REDIS
export REDIS_HOST=redis
export REDIS_PORT=6379
export REDIS_USERNAME=username
export REDIS_PASSWORD=password

#KIBANA
export ELASTICSEARCH_HOSTS=http://elasticsearch:9200
export LOGSTAH_HOST=logstash:12201
export KIBANA_HOST=kibana:5601

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