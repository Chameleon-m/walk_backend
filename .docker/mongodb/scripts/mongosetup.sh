#!/bin/bash

MONGODB1=mongo1

echo "**********************************************" ${MONGODB1}
echo "Waiting for startup.."
sleep 15
echo "done"
echo SETUP.sh time now: `date +"%T" `

mongosh --host ${MONGODB1}:27017 -u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD} <<EOF
var cfg = {
    "_id": "${MONGO_REPLICA_SET_NAME}",
    "members": [
        {
            "_id": 0,
            "host": "${MONGODB1}:27017"
        }
    ]
};
rs.initiate(cfg);
EOF

echo SETUP.sh END time now: `date +"%T" `