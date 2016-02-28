package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/russross/blackfriday"
)

// addr is server address.
const addr = ":8000"

const (
	indexTplFile = "templates/index.html"
	postTplFile  = "templates/post.html"
)

var (
	// database
	// db *sql.DB

	// templates
	indexTpl *template.Template
	postTpl  *template.Template
)

func init() {
	var err error
	/*
		// you do not have to open the db connection on every request
		// it can be done once at the start of the app
		db, err = sql.Open("mysql", "username:password(localhost:3306)/databasename")
		if err != nil {
			log.Fatal(err)
		}
		// Open doesn't open a connection. Validate DSN data:
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
	*/

	// init posts templates
	indexTpl, err = template.ParseFiles(indexTplFile)
	if err != nil {
		log.Fatal(err)
	}
	postTpl, err = template.ParseFiles(postTplFile)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	http.HandleFunc("/", handleRequest)
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("/location/onyourserver/css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("/location/onyourserver/js"))))

	log.Fatal(http.ListenAndServe(addr, nil))

}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	// skip if favicon.ico
	if r.URL.Path == "/favicon.ico" {
		return
	}

	/*
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
	*/
	if r.URL.Path[1:] == "" {
		posts, _ := Posts()
		if err := indexTpl.Execute(w, posts); err != nil {
			log.Print(err)
		}
		return
	}
	/*
		uniquepost := r.URL.Path[1:]
		rows, err := db.Query("select id, name, comment from comments where uniquepost = ?", uniquepost)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
	*/
	//declear an array to keep all comments
	var comments []Comment
	/*
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
	*/
	path := "posts/" + r.URL.Path[1:] + ".md"
	post, err := loadPost(path)
	if err != nil {
		log.Print(err)
		return
	}

	post.Comments = comments
	if err := postTpl.Execute(w, post); err != nil {
		log.Print(err)
	}
}

// Posts returns all posts from files.
func Posts() ([]*Post, error) {
	var all []*Post
	files, err := filepath.Glob("posts/*.md")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		p, err := loadPost(f)
		if err != nil {
			log.Print(err)
			continue
		}
		all = append(all, p)
	}
	// sort posts
	sort.Sort(byDate(all))
	return all, nil
}

func loadPost(path string) (*Post, error) {
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := newPost(fc)
	if err != nil {
		return nil, err
	}
	p.File = fileToken(path)
	return p, nil
}

// fileToken returns token from file path.
// Token is filename without extension.
func fileToken(path string) string {
	f := filepath.Base(path)
	ext := filepath.Ext(f)
	token := f[:len(f)-len(ext)]
	return token
}

// Post holds post data.
type Post struct {
	Title    string
	date     time.Time
	Summary  string
	Body     template.HTML
	File     string
	Comments []Comment
}

// Date returns formated date.
func (p *Post) Date() string {
	layout := "2 Jan 2006"
	return p.date.Format(layout)
}

const (
	// minLines holds minimal number of lines.
	minLines = 4

	titleLine   = 0
	dateLine    = 1
	summaryLine = 2
	bodyLine    = 3
)

func newPost(b []byte) (*Post, error) {
	if len(b) < 1 {
		return nil, errors.New("empty post")
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < minLines {
		return nil, errors.New("invalid post")
	}

	date, err := parseDate(lines[dateLine])
	if err != nil {
		return nil, err
	}

	body := strings.Join(lines[bodyLine:], "\n")

	p := &Post{
		Title:   lines[titleLine],
		date:    date,
		Summary: lines[summaryLine],
		Body: template.HTML(
			blackfriday.MarkdownCommon(
				[]byte(body),
			),
		),
	}

	return p, nil
}

// midnight returns date with zero time.
func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// iso8601DateFormat represents date in YYYY-MM-DD format.
const iso8601DateFormat = "2006-01-02"

// parseDate returns parsed or zero date.
func parseDate(s string) (time.Time, error) {
	d, err := time.Parse(iso8601DateFormat, s)
	return midnight(d), err
}

// byDate implements sort.Interface by providing
// Less and using the Len and Swap methods.
type byDate []*Post

// Len is length of posts.
func (bd byDate) Len() int {
	return len(bd)
}

// Less handle sort logic.
func (bd byDate) Less(i, j int) bool {
	return bd[i].date.After(bd[j].date)
}

// Swap changes elements.
func (bd byDate) Swap(i, j int) {
	bd[i], bd[j] = bd[j], bd[i]
}

// Comment holds comment data.
type Comment struct {
	Name, Comment string
}
