#!/bin/bash

docker run   --name=mongo -d -p 27017:27017 --network mongo -e  "MONGO_INITDB_ROOT_USERNAME=admin" -e "MONGO_INITDB_ROOT_PASSWORD=admin" mongo 

docker run   --name=mongo-express -d -p 8081:8081 --network mongo -e  "ME_MONGO_INITDB_ROOT_USERNAME=admin" -e "ME_MONGO_INITDB_ROOT_PASSWORD=admin" -e "ME_CONFIG_MONGODB_URL=mongodb://admin:admin@mongo:27017/" mongo-express