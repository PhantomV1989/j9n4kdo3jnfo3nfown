package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var (
	existingIds map[int]bool = map[int]bool{}
)

//sudo docker run --rm --network=host -e POSTGRES_PASSWORD=password -e POSTGRES_USER=user postgres

func initPostgres(ipPort string) *sql.DB {
	db, _ := sql.Open("postgres", "postgres://user:password@"+ipPort+"?sslmode=disable")
	_, err := db.Exec("CREATE DATABASE themoviedb") //"SELECT 1 FROM pg_database")
	if err != nil {
		log.Println(err.Error())
	}
	db.Close()
	db, _ = sql.Open("postgres", "postgres://user:password@"+ipPort+"/themoviedb?sslmode=disable")
	cstr := "CREATE TABLE movie_detail" +
		"( " +
		"    id integer CONSTRAINT movie_id PRIMARY KEY," +
		"    title varchar(200) NOT NULL," +
		"    popularity decimal," +
		"    overview  varchar(5000)," +
		"    vote_average decimal," +
		"    vote_count integer," +
		"    release_date date," +
		"    video boolean," +
		"    adult boolean," +
		"    original_language varchar(5)," +
		"    genre_ids integer[]" +
		");"
	_, err = db.Exec(cstr)
	if err != nil {
		log.Println(err.Error())
	}
	cstr = "CREATE TABLE top_rated" +
		"( " +
		"    rank_id integer CONSTRAINT rank_id PRIMARY KEY," +
		"    movie_id integer," +
		"    FOREIGN KEY (movie_id) REFERENCES movie_detail(id)" +
		");"
	_, err = db.Exec(cstr)
	if err != nil {
		log.Println(err.Error())
	}
	updateExistingMovieIDCache(db)
	return db
}

func updateExistingMovieIDCache(dbCon *sql.DB) {
	selDB, err := dbCon.Query("select id from movie_detail")
	if err != nil {
		print(err.Error())
	}
	for selDB.Next() {
		var id int
		err = selDB.Scan(&id)
		if err != nil {
			panic(err.Error())
		}
		existingIds[id] = true
	}
}

func searchTopRatedMovies(dbCon *sql.DB, n int, start, end string) []MovieDetail {
	/*	Only selected fields are queried. Can include full original fields if necessary
		select FIELDS from movie_detail, top_rated where
		top_rated.movie_id=movie_detail.id and
		movie_detail.release_date>='2020-01-01'
		order by rank_id asc limit 100;
	*/
	f := []string{
		"movie_detail.id",
		"movie_detail.title",
		"movie_detail.popularity",
		"movie_detail.overview",
		"movie_detail.release_date",
		"movie_detail.original_language",
		"movie_detail.adult",
	}
	q := "select " + strings.Join(f, ",") + " from movie_detail, top_rated where top_rated.movie_id=movie_detail.id and"
	if start != "" {
		q += " movie_detail.release_date>='" + start + "' and"
	}
	if end != "" {
		q += " movie_detail.release_date<='" + end + "'"
	}
	q = strings.TrimSuffix(q, "where")
	q = strings.TrimSuffix(q, "and")
	q += "order by rank_id asc limit " + strconv.Itoa(n) + ";"
	selDB, _ := dbCon.Query(q)
	results := []MovieDetail{}
	for selDB.Next() {
		_d := MovieDetail{}
		selDB.Scan(&_d.ID, &_d.Title, &_d.Popularity, &_d.Overview, &_d.ReleaseDate, &_d.OriginalLanguage, &_d.Adult)
		results = append(results, _d)
	}
	return results
}
