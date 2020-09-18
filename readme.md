# 1. Overview
This application is a standalone that serves top movie list from themoviedb.com

The application layer is written in Golang, using the Gin framework, while data is stored in a Postgres database.

On running the application, it will attempt to populate Postgresql with top handler_themoviedb.go/maxPageCount pages from themoviedb.com.


Once populated, the app will listen to port 8080 (defined in 2nd args), on route HTTP GET localhost:8080/v1/get_top_movies
Example:
```sh
curl localhost:8080/v1/get_top_movies?size=3&min_year=2015&max_year=2016
```
The fields min_year and max_year can be left out if desired.

JSON response is as follows:
```sh
{
    "results": [
        {
            "Popularity": 61.023,
            "Title": "Your Name.",
            "Overview": "High schoolers Mitsuha and Taki are complete strangers living separate lives. But one night, they suddenly switch places. Mitsuha wakes up in Taki’s body, and he in hers. This bizarre occurrence continues to happen randomly, and the two must adjust their lives around each other.",
            "vote_average": 0,
            "vote_count": 0,
            "release_date": "2016-08-26T00:00:00Z",
            "Video": false,
            "ID": 372058,
            "Adult": false,
            "original_language": "ja",
            "genre_ids": null
        },
		{...}
		...
    ]
}
```

# 2. Deployment
`Note: This whole setup was done in Ubuntu 18.04. You may use a similar *nix system`
For quickstart, you may run setup.sh.
## 2.1. Postgres
For convenience sake, run a Postgres container with the following command
```sh
docker run --rm --network=host -e POSTGRES_PASSWORD=password -e POSTGRES_USER=user postgres
```
## 2.2. Application
Run the following command to build the docker image.
```sh
sudo docker build --network=host -t belive .
```
Then start a container with the following:
```sh
sudo docker run --network=host belive localhost:5432 8080
```
The first argument "localhost:5432" refers to the service of the Postgres container started earlier. The 2nd argument "8080" refers to the port, the app is listening to.

# 3. Scaling
To scale the application, the API provider and the crawler has to be separated.

Step 1) The crawler will constantly retrieves the latest ranking results from themoviedb.com
Step 2) and updates the new data to Postgresql master.
Step 3) Postgresql master will update its read slaves for serving high volumes of read requests.
Step 4) The API workers will retrieve data from read slaves to serve clients
Step 5,6) A load balancer ensures that clients' requests are well distributed among API workers.

![alt text](https://github.com/PhantomV1989/j9n4kdo3jnfo3nfown/raw/master/deployment.jpg)
