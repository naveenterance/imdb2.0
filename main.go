package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-resty/resty/v2"
	_ "github.com/go-sql-driver/mysql"
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
var tpl *template.Template
var userID, hash string
var store = sessions.NewCookieStore([]byte("super-secret"))

func main() {
	tpl, _ = template.ParseGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:nst@tcp(localhost:3306)/recordings")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/loginauth", loginAuthHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerauth", registerAuthHandler)
	http.HandleFunc("/form", form)
	http.HandleFunc("/submit", processForm)
	http.HandleFunc("/data", data)
	http.HandleFunc("/ratings", ratings)
	// http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/", movie_list)

	http.ListenAndServe("localhost:8080", context.ClearHandler(http.DefaultServeMux))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****loginHandler running*****")
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****loginAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println("username:", username, "password:", password)

	stmt := "SELECT UserID, Hash FROM bcrypt WHERE Username = ?"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&userID, &hash)
	fmt.Println("hash from db:", hash)
	if err != nil {
		fmt.Println("error selecting Hash in db by Username")
		tpl.ExecuteTemplate(w, "login.html", "check username and password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {

		session, _ := store.Get(r, "session")

		session.Values["userID"] = userID

		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	fmt.Println("incorrect password")
	tpl.ExecuteTemplate(w, "login.html", "check username and password")
}

// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("*****indexHandler running*****")
// 	session, _ := store.Get(r, "session")
// 	_, ok := session.Values["userID"]
// 	fmt.Println("ok:", ok)
// 	if !ok {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// 	tpl.ExecuteTemplate(w, "index.html", "Logged In")
// }

// func aboutHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("*****aboutHandler running*****")
// 	session, _ := store.Get(r, "session")
// 	_, ok := session.Values["userID"]
// 	fmt.Println("ok:", ok)
// 	if !ok {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// 	tpl.ExecuteTemplate(w, "about.html", "Logged In")
// }

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****logoutHandler running*****")
	session, _ := store.Get(r, "session")

	delete(session.Values, "userID")
	session.Save(r, w)
	tpl.ExecuteTemplate(w, "login.html", "Logged Out")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****registerHandler running*****")
	tpl.ExecuteTemplate(w, "register.html", nil)
}

func registerAuthHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("*****registerAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")

	// var nameAlphaNumeric = true
	// for _, char := range username {

	// 	if unicode.IsLetter(char) == false && unicode.IsNumber(char) == false {
	// 		nameAlphaNumeric = false
	// 	}
	// }

	// var nameLength bool
	// if 5 <= len(username) && len(username) <= 50 {
	// 	nameLength = true
	// }

	password := r.FormValue("password")
	fmt.Println("password:", password, "\npswdLength:", len(password))

	// var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpaces bool
	// pswdNoSpaces = true
	// for _, char := range password {
	// 	switch {

	// 	case unicode.IsLower(char):
	// 		pswdLowercase = true

	// 	case unicode.IsUpper(char):
	// 		pswdUppercase = true

	// 	case unicode.IsNumber(char):
	// 		pswdNumber = true

	// 	case unicode.IsPunct(char) || unicode.IsSymbol(char):
	// 		pswdSpecial = true

	// 	case unicode.IsSpace(int32(char)):
	// 		pswdNoSpaces = false
	// 	}
	// }
	// if 11 < len(password) && len(password) < 60 {
	// 	pswdLength = true
	// }
	// fmt.Println("pswdLowercase:", pswdLowercase, "\npswdUppercase:", pswdUppercase, "\npswdNumber:", pswdNumber, "\npswdSpecial:", pswdSpecial, "\npswdLength:", pswdLength, "\npswdNoSpaces:", pswdNoSpaces, "\nnameAlphaNumeric:", nameAlphaNumeric, "\nnameLength:", nameLength)
	// if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpaces || !nameAlphaNumeric || !nameLength {
	// 	tpl.ExecuteTemplate(w, "register.html", "please check username and password criteria")
	// 	return
	// }

	stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		fmt.Println("username already exists, err:", err)
		tpl.ExecuteTemplate(w, "register.html", "username already taken")
		return
	}

	var hash []byte

	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Println("hash:", hash)
	fmt.Println("string(hash):", string(hash))

	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO bcrypt (Username, Hash) VALUES (?, ?);")
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	defer insertStmt.Close()
	var result sql.Result

	result, err = insertStmt.Exec(username, hash)
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	fmt.Println("rowsAff:", rowsAff)
	fmt.Println("lastIns:", lastIns)
	fmt.Println("err:", err)
	if err != nil {
		fmt.Println("error inserting new user")
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Fprint(w, "congrats, your account has been successfully created")
}

func movie_list(w http.ResponseWriter, r *http.Request) {

	fmt.Println("*****movie_list*****")
	session, _ := store.Get(r, "session")
	_, ok := session.Values["userID"]
	fmt.Println("ok:", ok)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	println("ID:", userID)
	name := userID
	rows, err := db.Query("SELECT * FROM movie_rating WHERE userID = ?", name)
	if err != nil {
		fmt.Printf("error")
		return
	}
	defer rows.Close()

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
		ratings := r.Form.Get("ratings")

		_, err = db.Exec("INSERT INTO movie_rating (userID, movieID, ratings) VALUES (?, ?, ?)", userID, IMDBID, ratings)
		if err != nil {
			fmt.Errorf("addAlbum: %v", err)
			return
		}

		fmt.Println("entered")
		return
	}

}
