package main

import (
	"database/sql"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetFormat struct {
	Size  int
	Start string
	End   string
}

var (
	dbCon = &sql.DB{}
)

func getTopMovies(c *gin.Context) {
	//localhost:8080/v1/get_top_movies?size=3&min_year=2015&max_year=2016
	//start and end are optional
	size := c.Query("size")
	if size == "" {
		size = "10"
	}
	sizei, err := strconv.Atoi(size)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "size should be an integer.",
		})
	} else {
		start := c.Query("min_year")
		end := c.Query("max_year")
		if start != "" {
			start += "-01-01"
		}
		if end != "" {
			end += "-12-31"
		}
		results := searchTopRatedMovies(dbCon, sizei, start, end)
		c.JSON(200, gin.H{
			"results": results,
		})
	}
}

func main() {
	println(os.Args[1])
	dbCon = initPostgres(os.Args[1])
	updateDatabaseFromMovieDB(dbCon)
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.GET("/get_top_movies", getTopMovies)
	}
	router.Run(":" + os.Args[2])
}
