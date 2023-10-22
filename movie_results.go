

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/go-resty/resty/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

type SearchResult struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	IMDBID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

type SearchResults struct {
	Search       []SearchResult `json:"Search"`
	TotalResults string         `json:"totalResults"`
	Response     string         `json:"Response"`
}

type Films struct {
	UserID  string
	MovieID string
	Ratings string
}

type Movie struct {
	Title string `json:"Title"`
	Year  string `json:"Year"`
}

type FilmsInfo struct {
	Title   string
	Year    string
	Ratings string
}

var movinfo FilmsInfo
var movinfos []FilmsInfo
var searchTerm string
var db *sql.DB
var store = sessions.NewCookieStore([]byte("secret-key"))

func form(w http.ResponseWriter, r *http.Request) {

	fmt.Println("form vewing")

	tmpl, err := template.ParseFiles("templates/search_movies.html")
	if err != nil {
		fmt.Println("Index Template Parse Error: ", err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Index Template Execution Error: ", err)
	}
}

func processForm(w http.ResponseWriter, r *http.Request) {

	fmt.Println("form processing")
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read form data
	r.ParseForm()
	searchTerm = r.FormValue("name")

	http.Redirect(w, r, "/data", http.StatusSeeOther)

}

func data(w http.ResponseWriter, r *http.Request) {
	fmt.Println("results")

	baseURL := "http://www.omdbapi.com/"

	apiKey := "e24ea998"

	urlStr := fmt.Sprintf("%s?apikey=%s&s=%s", baseURL, apiKey, url.QueryEscape(searchTerm))

	resp, err := http.Get(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var searchResults SearchResults
	err = json.NewDecoder(resp.Body).Decode(&searchResults)
	if err != nil {
		log.Fatal(err)
	}
	tmpl := template.Must(template.ParseFiles("templates/movie_results.html"))

	fmt.Println("Search Results:")
	for _, result := range searchResults.Search {
		fmt.Printf("%s (%s) %s \n", result.Title, result.Year, result.IMDBID)
		tmpl.Execute(w, result)

	}

}
func ratings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "Error parsing form data")
			return
		}

		IMDBID := r.Form.Get("IMDBID")
		userID := 1
		ratings := 5

		_, err = db.Exec("INSERT INTO movie_rating (userID, movieID, ratings) VALUES (?, ?, ?)", userID, IMDBID, ratings)
		if err != nil {
			fmt.Errorf("addAlbum: %v", err)
			return
		}

		fmt.Println("entered")
		return
	}

}

func movie_list(w http.ResponseWriter, r *http.Request) {

	name := 2
	rows, err := db.Query("SELECT * FROM movie_rating WHERE userID = ?", name)
	if err != nil {
		fmt.Printf("error")
		return
	}
	defer rows.Close()

	session, _ := store.Get(r, "session-name")

	if session.Values["once"] != true {

		for rows.Next() {
			var mov Films
			if err := rows.Scan(&mov.UserID, &mov.MovieID, &mov.Ratings); err != nil {
				fmt.Printf("error")
			}

			movie, err := getMovieByID(mov.MovieID)
			if err != nil {
				log.Fatalf("failed to get movie details: %v", err)
			}

			fmt.Printf("Movie: %s\nYear: %s\n", movie.Title, movie.Year)
			movinfo.Title = movie.Title
			movinfo.Year = movie.Year
			movinfo.Ratings = mov.Ratings

			movinfos = append(movinfos, movinfo)

		}
		session.Values["once"] = true
		session.Save(r, w)
	}

	fmt.Println(movinfos)

	tmpl := template.Must(template.ParseFiles("templates/movie_list.html"))
	err = tmpl.Execute(w, movinfos)
	if err != nil {
		fmt.Println(err)
		return

	}
}
func getMovieByID(id string) (*Movie, error) {
	client := resty.New()
	response, err := client.R().
		SetQueryParams(map[string]string{
			"i": id,
		}).
		SetResult(&Movie{}).
		Get("http://www.omdbapi.com/?apikey=e24ea998")

	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return nil, fmt.Errorf("failed to get movie details: %v", response.Error())
	}

	movie := response.Result().(*Movie)
	return movie, nil
}

func main() {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "nst",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "recordings",
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	http.HandleFunc("/", movie_list)
	http.HandleFunc("/form", form)
	http.HandleFunc("/submit", processForm)
	http.HandleFunc("/data", data)
	http.HandleFunc("/ratings", ratings)
	http.ListenAndServe(":8080", nil)

}
