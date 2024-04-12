package dbsubs

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"

	_ "github.com/lib/pq"
)

func GetDbUrl() string {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	return dbUrl
}

type SubWithDB struct {
	subs map[*websocket.Conn]bool
	db   *sql.DB
}

func RanHash() string {
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var slug string

	for i := 0; i < 8; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			fmt.Println("Error generating random index:", err)
			return ""
		}
		slug += string(characters[randomIndex.Int64()])
	}

	return slug
}

func NewDb() *sql.DB {
	db, err := sql.Open("postgres", GetDbUrl())
	if err != nil {
		fmt.Println("db conn err: ", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS subs(
		id TEXT PRIMARY KEY,
		name TEXT,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS notifs(
		id TEXT PRIMARY KEY,
		title TEXT,
		content TEXT,
		valid_till TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW()
	)
	`

	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("create table err: ", err)
	}

	return db
}

func (s *SubWithDB) insertSub(name string) (string, error) {
	query := `
	INSERT INTO subs(id, name, password)
	VALUES ($1, $2, $3)
	`

	password := RanHash()
	_, err := s.db.Exec(query, RanHash(), name, password)
	if err != nil {
		return "", err
	}
	return password, nil
}

func (s *SubWithDB) getName(password string) (string, error) {
	query := `
	SELECT * from subs where password=$1;
	`

	var name string
	rows := s.db.QueryRow(query, password)
	err := rows.Scan(nil, &name, nil, nil)
	if err != nil {
		return "", err
	}
	return name, nil
}

func NewSubWithDB() *SubWithDB {
	return &SubWithDB{
		subs: make(map[*websocket.Conn]bool),
		db:   NewDb(),
	}
}

func (s *SubWithDB) broadcast(msg string) {
	for ws, stat := range s.subs {
		if stat {
			go func(c *websocket.Conn) {
				if _, err := c.Write([]byte(msg)); err != nil {
					fmt.Println("Write Error: ", err)
				}
			}(ws)
		}
	}
}

func (s *SubWithDB) handleSubscriber(ws *websocket.Conn) {
	for {
		headers := ws.Request().Header
		auth_header := headers.Get("Authorization")
		if auth_header == "" {
			ws.Write([]byte("No auth header found"))
		}
		_, err := s.getName(auth_header)
		if err != nil {
			fmt.Println("No name with auth header: ", auth_header, " failed with err: ", err)
		}
		ws.Write([]byte("Listening :)"))
	}
}

func (s *SubWithDB) handleBroadCast(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println("FormParse err: ", err)
		}
		msg := r.Form.Get("msg")
		if msg == "" {
			http.Error(w, "msg parameter missing", http.StatusBadRequest)
			return
		}

		s.broadcast(msg)
	} else {
		fmt.Fprintf(w, "Post only :)")
	}
}

func (s *SubWithDB) handleSub(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println("FormParse err: ", err)
		}
		name := r.Form.Get("name")
		if name == "" {
			http.Error(w, "Name parameter missing", http.StatusBadRequest)
			return
		}

		password, err := s.insertSub(name)
		if err != nil {
			fmt.Println("insert err: ", err)
			return
		}

		fmt.Fprintf(w, "Subscribed %s to the feed with password %s", name, password)
	} else {
		w.Write([]byte("Post only :)"))
	}
}

func (s *SubWithDB) StartNewSubWithDBServer(port string) {
	http.HandleFunc("/sub", s.handleSub)
	http.Handle("/listen", websocket.Handler(s.handleSubscriber))
	http.ListenAndServe(port, nil)
}
