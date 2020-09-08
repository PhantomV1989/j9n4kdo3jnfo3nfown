sudo docker run -d --rm --network=host -e POSTGRES_PASSWORD=password -e POSTGRES_USER=user postgres && \
sudo docker build --network=host -t belive . && \
sudo docker run --network=host belive lochost:5432 8080

