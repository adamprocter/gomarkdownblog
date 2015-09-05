package main

import (
	//"fmt"
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/russross/blackfriday"
)


type Post struct {
	Status   string
	Title    string
	Date     string
	Summary  string
	Body     template.HTML
	File     string
	Comments []Comment
}

type Comment struct {
	Name, Comment string
}

var db *sql.DB

func init() {
	// you do not have to open the db connection on every request
	// it can be done once at the start of the app
	var err error
	db, err = sql.Open("mysql", "username:password(localhost:3306)/databasename")
	if err != nil {
		log.Fatal(err)
	}
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

}

func handlerequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		uniquepost := r.FormValue("uniquepost")
		namein := r.FormValue("name")
		commentin := r.FormValue("comment")

		_, err := db.Exec(
			"INSERT INTO comments (uniquepost, name, comment) VALUES (?, ?, ?)",
			uniquepost,
			namein,
			commentin,
		)
		if err != nil {
			log.Fatal(err)
		}
		//when done inserting comment redirect back to this page
		http.Redirect(w, r, r.URL.Path, 301)
		return
	}

	if r.URL.Path[1:] == "" {
		posts := getPosts()
		t := template.New("index.html")
		t, _ = t.ParseFiles("index.html")
		t.Execute(w, posts)
		return
	}
	uniquepost := r.URL.Path[1:]
	rows, err := db.Query("select id, name, comment from comments where uniquepost = ?", uniquepost)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//declear an array to keep all comments
	var comments []Comment

	for rows.Next() {
		var id int
		var name, comment string
		err := rows.Scan(&id, &name, &comment)
		if err != nil {
			log.Fatal(err)
		}
		//append the comment into the array when done
		comments = append(comments, Comment{name, comment})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	f := "posts/" + r.URL.Path[1:] + ".md"
	fileread, _ := ioutil.ReadFile(f)
	lines := strings.Split(string(fileread), "\n")
	status := string(lines[0])
	title := string(lines[1])
	date := string(lines[2])
	summary := string(lines[3])
	body := strings.Join(lines[4:len(lines)], "\n")
	htmlBody := template.HTML(blackfriday.MarkdownCommon([]byte(body)))
	post := Post{status,title, date, summary, htmlBody, r.URL.Path[1:], comments}
	t := template.New("post.html")
	t, _ = t.ParseFiles("post.html")
	t.Execute(w, post)

}

func getPosts() []Post {
	a := []Post{}
	files, _ := filepath.Glob("posts/*")
	for _, f := range files {
		file := strings.Replace(f, "posts/", "", -1)
		file = strings.Replace(file, ".md", "", -1)
		fileread, _ := ioutil.ReadFile(f)
		lines := strings.Split(string(fileread), "\n")
		status := string(lines[0])
		title := string(lines[1])
		date := string(lines[2])
		summary := string(lines[3])
		body := strings.Join(lines[4:len(lines)], "\n")
		htmlBody := template.HTML(blackfriday.MarkdownCommon([]byte(body)))

		a = append(a, Post{status, title, date, summary, htmlBody, file, nil})
	}
	return a
}

func main() {
http.HandleFunc("/", handlerequest)
http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("/location/onyourserver/css"))))
http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("/location/onyourserver/js"))))

	http.ListenAndServe(":8000", nil)

}
