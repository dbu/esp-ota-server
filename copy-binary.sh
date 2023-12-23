id=$(docker create esp-ota-server)
docker cp $id:espotad .
docker rm -v $id
