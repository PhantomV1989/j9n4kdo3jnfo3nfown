package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	dbCon := initPostgres("localhost:5432")
	defer dbCon.Close()
	n := 5
	res := searchTopRatedMovies(dbCon, n, "", "")
	assert.Len(t, res, n)

	n = 10
	res = searchTopRatedMovies(dbCon, n, "2020-01-01", "")
	for d := range res {
		assert.True(t, strings.Contains(res[d].ReleaseDate, "2020"))
	}
	assert.Len(t, res, n)

	n = 8
	res = searchTopRatedMovies(dbCon, n, "2015-01-01", "2015-12-31")
	for d := range res {
		assert.True(t, strings.Contains(res[d].ReleaseDate, "2015"))
	}
	assert.Len(t, res, n)
}

func TestUpdateDatabase(t *testing.T) {
	dbCon := initPostgres("localhost:5432")
	defer dbCon.Close()
	updateDatabaseFromMovieDB(dbCon)
}

func TestResponseParsers(t *testing.T) {
	testString := "{\"page\":1,\"total_results\":7762,\"total_pages\":389,\"results\":[{\"popularity\":18.176,\"vote_count\":743,\"video\":false,\"poster_path\":\"/a.jpg\",\"id\":724089,\"adult\":false,\"backdrop_path\":\"/b.jpg\",\"original_language\":\"en\",\"original_title\":\"Gabriel's Inferno Part II\",\"genre_ids\":[10749],\"title\":\"Gabriel's Inferno Part II\",\"vote_average\":9.1,\"overview\":\"Professor Gabriel Emerson...\",\"release_date\":\"2020-07-31\"},{\"popularity\":14.497,\"vote_count\":1331,\"video\":false,\"poster_path\":\"/q.jpg\",\"id\":696374,\"adult\":false,\"backdrop_path\":\"/b.jpg\",\"original_language\":\"en\",\"original_title\":\"Gabriel's Inferno\",\"genre_ids\":[10749],\"title\":\"Gabriel's Inferno\",\"vote_average\":9,\"overview\":\"An intriguing...\",\"release_date\":\"2020-05-29\"}]}"

	var body map[string]interface{}
	json.Unmarshal([]byte(testString), &body)

	pageCount, err := getTotalPageCount(body)
	assert.Empty(t, err)
	assert.Equal(t, pageCount, 389)

	movieDetails, err := getMovieDetailList(body)
	assert.Empty(t, err)
	assert.Equal(t, len(movieDetails), 2)
	assert.Equal(t, movieDetails[0].Title, "Gabriel's Inferno Part II")
}

func TestMovieDetailParser(t *testing.T) {
	testString := "{\"popularity\": 18.176,\"vote_count\": 743,\"video\": false,\"poster_path\": \"/pci1ArYW7oJ2eyTo2NMYEKHHiCP.jpg\",\"id\": 724089,\"adult\": false,\"backdrop_path\": \"/jtAI6OJIWLWiRItNSZoWjrsUtmi.jpg\",\"original_language\": \"en\",\"original_title\": \"Gabriel's Inferno Part II\",\"genre_ids\": [10749],\"title\": \"Gabriel's Inferno Part II\",\"vote_average\": 9.1,\"overview\": \"Professor Gabriel Emerson finally learns the truth about Julia Mitchell's identity, but his realization comes a moment too late. Julia is done waiting for the well-respected Dante specialist to remember her and wants nothing more to do with him. Can Gabriel win back her heart before she finds love in another's arms?\",\"release_date\": \"2020-07-31\"}"
	movieDetail, err := parseMovieDetail(testString)
	assert.Empty(t, err)
	assert.Equal(t, movieDetail.Popularity, float32(18.176))
	assert.Equal(t, movieDetail.VoteCount, 743)
	assert.Equal(t, movieDetail.ReleaseDate, "2020-07-31")
}
