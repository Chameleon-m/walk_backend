#!/bin/bash
echo "START create user..."
mongosh -u "$MONGO_INITDB_ROOT_USERNAME" -p "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin "$MONGO_INITDB_NAME" <<EOF
    db.createUser({ user: '$MONGO_INITDB_USERNAME', pwd: '$MONGO_INITDB_PASSWORD', roles: [ { role: 'dbOwner', db: '$MONGO_INITDB_NAME' } ] })
EOF
echo "END create user."