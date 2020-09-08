package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"
)

/*
	API keys
	11004c5dda64d0bae607c7af2636e983
	85f860ff9df4b2cc8f19244ab333c53b
*/

// MovieDetail ..
type MovieDetail struct {
	Popularity       float32
	Title            string
	Overview         string
	VoteAverage      float32 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	ReleaseDate      string  `json:"release_date"`
	Video            bool
	ID               int
	Adult            bool
	OriginalLanguage string `json:"original_language"`
	GenreIDs         []int  `json:"genre_ids"`
}

var (
	maxPageCount = 10 // 20 items * n pages
)

func updateDatabaseFromMovieDB(dbCon *sql.DB) error {
	//https://api.themoviedb.org/3/movie/top_rated?api_key=11004c5dda64d0bae607c7af2636e983&page=1
	// Becareful of deadlocks, reads while updating DB
	log.Println("Updating database, please wait...")
	client := &http.Client{}
	rankID := 0

	processPage := func(pageNo int) (int, error) {
		req, err := http.NewRequest("GET", "https://api.themoviedb.org/3/movie/top_rated?api_key=11004c5dda64d0bae607c7af2636e983&page="+strconv.Itoa(pageNo), strings.NewReader(""))
		if err != nil {
			return -1, err
		}
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return -1, err
		}
		sbody, err := ioutil.ReadAll(resp.Body)

		var body map[string]interface{}
		err = json.Unmarshal(sbody, &body)
		if err != nil {
			return -1, err
		}
		totalPages, err := getTotalPageCount(body)
		if err != nil {
			return -1, err
		}
		movies, err := getMovieDetailList(body)
		if err != nil {
			return -1, err
		}
		if totalPages > maxPageCount {
			totalPages = maxPageCount
		}
		updateMovieDetailsTable(movies, dbCon)
		updateTopRatedTable(movies, dbCon, &rankID)
		log.Println("Updated page " + strconv.Itoa(pageNo) + " of " + strconv.Itoa(totalPages))
		return totalPages, nil
	}
	totalPages, err := processPage(1)
	if err != nil {
		return err
	}

	for currentPage := 2; currentPage < totalPages; currentPage++ {
		processPage(currentPage)
	}
	return nil
}

func parseMovieDetail(s string) (MovieDetail, error) {
	md := MovieDetail{}
	err := json.Unmarshal([]byte(s), &md)
	return md, err
}

func getTotalPageCount(responseBody map[string]interface{}) (int, error) {
	totalPages := -1
	if _, exists := responseBody["total_pages"]; exists {
		totalPages = int(responseBody["total_pages"].(float64))
		return totalPages, nil
	}
	return totalPages, errors.New("total_pages not found in response")
}

func getMovieDetailList(responseBody map[string]interface{}) ([]MovieDetail, error) {
	results := []MovieDetail{}
	if _, exists := responseBody["results"]; exists {
		rarr := responseBody["results"].([]interface{})
		for i := range rarr {
			md := MovieDetail{}
			// Marshal/Unmarshal to remap. Might be unoptimised
			_b, _ := json.Marshal(rarr[i])
			json.Unmarshal(_b, &md)
			results = append(results, md)
		}
		return results, nil
	}
	return results, errors.New("results not found in response")
}

func updateMovieDetailsTable(movieDetails []MovieDetail, dbCon *sql.DB) {
	toInsertString := func(md MovieDetail) string {
		parseBool := func(b bool) string {
			if b {
				return "TRUE"
			}
			return "FALSE"
		}
		gids := ""
		for i := range md.GenreIDs {
			gids += strconv.Itoa(md.GenreIDs[i]) + ","
		}
		if len(md.GenreIDs) > 0 {
			gids = "[" + gids[:len(gids)-1] + "]"
		} else {
			gids = "[]::int[]"
		}

		values := []string{
			strconv.Itoa(md.ID),
			"e'" + strings.ReplaceAll(md.Title, "'", "''") + "'",
			fmt.Sprintf("%f", md.Popularity),
			"e'" + strings.ReplaceAll(md.Overview, "'", "''") + "'",
			fmt.Sprintf("%f", md.VoteAverage),
			strconv.Itoa(md.VoteCount),
			"'" + md.ReleaseDate + "'",
			parseBool(md.Video),
			parseBool(md.Adult),
			"'" + md.OriginalLanguage + "'",
			"ARRAY " + gids,
		}
		return "(" + strings.Join(values, ",") + ")"
	}
	q := "insert into movie_detail(id,title,popularity,overview,vote_average,vote_count,release_date,video,adult,original_language,genre_ids) values "
	for mids := range movieDetails {
		md := movieDetails[mids]
		if _, exists := existingIds[md.ID]; !exists {
			q += toInsertString(md) + ","
			existingIds[md.ID] = true
		}
	}

	q = q[:len(q)-1]
	if string(q[len(q)-1]) == ")" {
		_, err := dbCon.Exec(q)
		if err != nil {
			log.Println(err.Error())
			log.Println(q)
		}
	}
}

func updateTopRatedTable(movieDetails []MovieDetail, dbCon *sql.DB, rankid *int) {
	q := "insert into top_rated(rank_id,movie_id) values "
	for mids := range movieDetails {
		md := movieDetails[mids]
		q += "(" + strconv.Itoa(*rankid) + "," + strconv.Itoa(md.ID) + "),"
		*rankid++
	}
	q = q[:len(q)-1]
	if len(movieDetails) > 0 {
		_, err := dbCon.Exec(q + " ON CONFLICT (rank_id) DO UPDATE SET movie_id = excluded.movie_id")
		if err != nil {
			log.Println(err.Error())
			log.Println(q)
		}
	}
}
