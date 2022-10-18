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

# GIN
export PORT=8080
export GIN_MODE=debug
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