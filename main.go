package main

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DATABASE_PROVIDER = "sqlite3"
	DATABASE_NAME     = "./local.db"
)

var (
	db *sql.DB
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserSubject struct {
	ID       int       `json:"id"`
	Subjects []Subject `json:"subjects"`
}

type Subject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Booth struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Content      string `json:"content"`
	PlayingUsers []User `json:"playing_users"`
}

type Score struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	BoothID   int    `json:"booth_id"`
	Score     int    `json:"score"`
	CreatedAt string `json:"created_at"`
}

func GetRouter() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		user := api.Group("/user")
		{
			user.GET("/:user_id/total_score", GetUserTotalScoreRouter)
			user.GET("/:user_id/exist", CheckUserExistRouter)

			user.POST("/", RegisterUserRouter)
		}

		booth := api.Group("/booth")
		{
			booth.GET("/", GetBoothsRouter)
			booth.GET("/:booth_id", GetBoothRouter)

			booth.POST("/new", NewBoothRouter)
			booth.POST("/edit", EditBoothRouter)
		}

		subject := api.Group("/subject")
		{
			subject.GET("/", GetSubjectsRouter)
			subject.GET("/:user_id", GetUserSubjectsRouter)

			subject.POST("/:user_id", UpdateUserSubjectsRouter)
		}

		score := api.Group("/score")
		{
			score.POST("/", NewScoreRouter)
			score.GET("/:user_id", GetUserScoresRouter)
			score.GET("/:user_id/total_score", GetUserTotalScoreRouter)
		}
	}

	return router
}

func SetupDatabase() {
	db, _ = sql.Open(DATABASE_PROVIDER, DATABASE_NAME)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, password TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS subjects (id INTEGER PRIMARY KEY, name TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS user_subjects (user_id INTEGER, subject_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS booths (id INTEGER PRIMARY KEY, name TEXT, content TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS scores (id INTEGER PRIMARY KEY, user_id INTEGER, booth_id INTEGER, score INTEGER, created_at TEXT)")
}

func CheckUserExistRouter(c *gin.Context) {
	userID := c.Param("user_id")
	userIDInt, _ := strconv.Atoi(userID)

	var exist bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userIDInt).Scan(&exist)

	c.JSON(200, gin.H{"exist": exist})
}

func RegisterUserRouter(c *gin.Context) {
	var user User
	c.BindJSON(&user)

	db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)

	c.JSON(200, gin.H{"id": user.ID})
}

func NewScoreRouter(c *gin.Context) {
	var score Score
	c.BindJSON(&score)

	currentTime := time.Now()
	score.CreatedAt = currentTime.Format(time.RFC1123)

	db.Exec("INSERT INTO scores (user_id, booth_id, score, created_at) VALUES (?, ?, ?, ?)", score.UserID, score.BoothID, score.Score, score.CreatedAt)

	c.JSON(200, gin.H{"id": score.ID})
}

func GetUserScores(userID int) []Score {
	rows, _ := db.Query("SELECT * FROM scores WHERE user_id = ?", userID)

	var scores []Score
	for rows.Next() {
		var score Score
		rows.Scan(&score.ID, &score.UserID, &score.BoothID, &score.Score, &score.CreatedAt)
		scores = append(scores, score)
	}

	return scores
}

func GetUserTotalScoreRouter(c *gin.Context) {
	userID := c.Param("user_id")
	userIDInt, _ := strconv.Atoi(userID)

	scores := GetUserScores(userIDInt)
	totalScore := 0
	for _, score := range scores {
		totalScore += score.Score
	}

	c.JSON(200, gin.H{"total_score": totalScore})
}

func GetUserScoresRouter(c *gin.Context) {
	userID := c.Param("user_id")
	userIDInt, _ := strconv.Atoi(userID)

	scores := GetUserScores(userIDInt)

	c.JSON(200, gin.H{"scores": scores})
}

func UpdateUserSubjectsRouter(c *gin.Context) {
	userID := c.Param("user_id")
	userIDInt, _ := strconv.Atoi(userID)

	var userSubject UserSubject
	c.BindJSON(&userSubject)

	db.Exec("DELETE FROM user_subjects WHERE user_id = ?", userIDInt)

	for _, subject := range userSubject.Subjects {
		db.Exec("INSERT INTO user_subjects (user_id, subject_id) VALUES (?, ?)", userSubject.ID, subject.ID)
	}

	c.JSON(200, gin.H{"id": userSubject.ID})
}

func GetUserSubjectsRouter(c *gin.Context) {
	userID := c.Param("user_id")
	userIDInt, _ := strconv.Atoi(userID)

	rows, _ := db.Query("SELECT * FROM user_subjects WHERE user_id = ?", userIDInt)

	var userSubject UserSubject
	userSubject.ID = userIDInt
	for rows.Next() {
		var subject Subject
		rows.Scan(&subject.ID)
		userSubject.Subjects = append(userSubject.Subjects, subject)
	}

	c.JSON(200, gin.H{"subjects": userSubject.Subjects})
}

func NewBoothRouter(c *gin.Context) {
	var booth Booth
	c.BindJSON(&booth)

	db.Exec("INSERT INTO booths (name, content) VALUES (?, ?)", booth.Name, booth.Content)

	c.JSON(200, gin.H{"id": booth.ID})
}

func EditBoothRouter(c *gin.Context) {
	var booth Booth
	c.BindJSON(&booth)

	db.Exec("UPDATE booths SET name = ?, content = ? WHERE id = ?", booth.Name, booth.Content, booth.ID)

	c.JSON(200, gin.H{"id": booth.ID})
}

func GetBoothRouter(c *gin.Context) {
	boothID := c.Param("booth_id")
	boothIDInt, _ := strconv.Atoi(boothID)

	rows, _ := db.Query("SELECT * FROM booths WHERE id = ?", boothIDInt)

	var booth Booth
	for rows.Next() {
		rows.Scan(&booth.ID, &booth.Name, &booth.Content)
	}

	c.JSON(200, gin.H{"booth": booth})
}

func GetBoothsRouter(c *gin.Context) {
	rows, _ := db.Query("SELECT * FROM booths")

	var booths []Booth
	for rows.Next() {
		var booth Booth
		rows.Scan(&booth.ID, &booth.Name, &booth.Content)
		booths = append(booths, booth)
	}

	c.JSON(200, gin.H{"booths": booths})
}

func GetSubjectsRouter(c *gin.Context) {
	rows, _ := db.Query("SELECT * FROM subjects")

	var subjects []Subject
	for rows.Next() {
		var subject Subject
		rows.Scan(&subject.ID, &subject.Name)
		subjects = append(subjects, subject)
	}

	c.JSON(200, gin.H{"subjects": subjects})
}

func main() {
	SetupDatabase()

	router := GetRouter()
	router.Run(":8080")
}
