package main

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
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
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserSubject struct {
	ID       string    `json:"id"`
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

type BoothImage struct {
	ID      int    `json:"id"`
	BoothID int    `json:"booth_id"`
	Image   string `json:"image"`
}

type BoothVideo struct {
	ID      int    `json:"id"`
	BoothID int    `json:"booth_id"`
	URL     string `json:"url"`
}

type BoothBook struct {
	ID      int    `json:"id"`
	BoothID int    `json:"booth_id"`
	UserID  string `json:"user_id"`
	Period  int    `json:"period"`
}

type Score struct {
	ID        int    `json:"id"`
	UserID    string `json:"user_id"`
	BoothID   int    `json:"booth_id"`
	Score     int    `json:"score"`
	CreatedAt string `json:"created_at"`
}

type BoothBookUser struct {
	userID string `json:"user_id"`
}

func GetRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

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
			booth.GET("/:booth_id/image", GetBoothImagesRouter)
			booth.GET("/:booth_id/video", GetBoothVideosRouter)

			booth.POST("/new", NewBoothRouter)
			booth.POST("/edit", EditBoothRouter)
			booth.POST("/image", NewBoothImageRouter)
			booth.POST("/video", NewBoothVideoRouter)

			book := booth.Group("/book")
			{
				book.POST("/:booth_id/:period", NewBoothBookRouter)
				book.GET("/:booth_id", GetBoothBooksRouter)
				book.GET("/:booth_id/:period", GetBoothBookRouter)
				book.GET("/u/:user_id", GetUserBooksRouter)
				book.DELETE("/:booth_id/:period", DeleteBoothBookRouter)
			}
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

	db.Exec("CREATE TABLE IF NOT EXISTS users (id TEXT, username TEXT, password TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS subjects (id INTEGER PRIMARY KEY, name TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS user_subjects (user_id TEXT, subject_id INTEGER)")
	db.Exec("CREATE TABLE IF NOT EXISTS booths (id INTEGER PRIMARY KEY, name TEXT, content TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS scores (id INTEGER PRIMARY KEY, user_id TEXT, booth_id INTEGER, score INTEGER, created_at TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS booth_images (id INTEGER PRIMARY KEY, booth_id INTEGER, image TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS booth_videos (id INTEGER PRIMARY KEY, booth_id INTEGER, url TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS booth_books (id INTEGER PRIMARY KEY, booth_id INTEGER, user_id TEXT, period INTEGER)")
}

func GetBoothVideosRouter(c *gin.Context) {
	boothID := c.Param("booth_id")

	var videos []BoothVideo
	rows, _ := db.Query("SELECT * FROM booth_videos WHERE booth_id = ?", boothID)
	for rows.Next() {
		var video BoothVideo
		rows.Scan(&video.ID, &video.BoothID, &video.URL)
		videos = append(videos, video)
	}

	c.JSON(200, gin.H{"videos": videos})
}

func NewBoothVideoRouter(c *gin.Context) {
	var video BoothVideo
	c.BindJSON(&video)

	db.Exec("INSERT INTO booth_videos (booth_id, url) VALUES (?, ?)", video.BoothID, video.URL)

	c.JSON(200, gin.H{"id": video.ID})
}

func GetBoothImagesRouter(c *gin.Context) {
	boothID := c.Param("booth_id")

	var images []BoothImage
	rows, _ := db.Query("SELECT * FROM booth_images WHERE booth_id = ?", boothID)
	for rows.Next() {
		var image BoothImage
		rows.Scan(&image.ID, &image.BoothID, &image.Image)
		images = append(images, image)
	}

	c.JSON(200, gin.H{"images": images})
}

func NewBoothImageRouter(c *gin.Context) {
	var image BoothImage
	c.BindJSON(&image)

	db.Exec("INSERT INTO booth_images (booth_id, image) VALUES (?, ?)", image.BoothID, image.Image)

	c.JSON(200, gin.H{"id": image.ID})
}

func CheckUserExistRouter(c *gin.Context) {
	userID := c.Param("user_id")

	var exist bool
	db.QueryRow("SELECT EXISTS(SELECT * FROM users WHERE id = ?)", userID).Scan(&exist)

	c.JSON(200, gin.H{"exist": exist})
}

func RegisterUserRouter(c *gin.Context) {
	var user User
	c.BindJSON(&user)

	db.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)", user.ID, user.Username, user.Password)

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

func GetUserScores(userID string) []Score {
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

	scores := GetUserScores(userID)
	totalScore := 0
	for _, score := range scores {
		totalScore += score.Score
	}

	c.JSON(200, gin.H{"total_score": totalScore})
}

func GetUserScoresRouter(c *gin.Context) {
	userID := c.Param("user_id")

	scores := GetUserScores(userID)

	c.JSON(200, gin.H{"scores": scores})
}

func UpdateUserSubjectsRouter(c *gin.Context) {
	userID := c.Param("user_id")

	var userSubject UserSubject
	c.BindJSON(&userSubject)

	db.Exec("DELETE FROM user_subjects WHERE user_id = ?", userID)

	for _, subject := range userSubject.Subjects {
		db.Exec("INSERT INTO user_subjects (user_id, subject_id) VALUES (?, ?)", userSubject.ID, subject.ID)
	}

	c.JSON(200, gin.H{"id": userSubject.ID})
}

func GetUserSubjectsRouter(c *gin.Context) {
	userID := c.Param("user_id")

	rows, _ := db.Query("SELECT * FROM user_subjects WHERE user_id = ?", userID)

	var userSubject UserSubject
	userSubject.ID = userID
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

func NewBoothBookRouter(c *gin.Context) {
	boothID := c.Param("booth_id")
	period := c.Param("period")

	var respData BoothBookUser
	c.BindJSON(&respData)
	userID := respData.userID

	db.Exec("INSERT INTO booth_books (booth_id, user_id, period) VALUES (?, ?, ?)", boothID, userID, period)

	c.JSON(200, gin.H{"id": userID})
}

func GetBoothBooksRouter(c *gin.Context) {
	boothID := c.Param("booth_id")

	rows, _ := db.Query("SELECT * FROM booth_books WHERE booth_id = ?", boothID)

	var boothBooks []BoothBook
	for rows.Next() {
		var boothBook BoothBook
		rows.Scan(&boothBook.ID, &boothBook.BoothID, &boothBook.UserID, &boothBook.Period)
		boothBooks = append(boothBooks, boothBook)
	}

	c.JSON(200, gin.H{"booth_books": boothBooks})
}

func GetBoothBookRouter(c *gin.Context) {
	boothID := c.Param("booth_id")
	period := c.Param("period")

	rows, _ := db.Query("SELECT * FROM booth_books WHERE booth_id = ? AND period = ?", boothID, period)

	var boothBooks []BoothBook
	for rows.Next() {
		var boothBook BoothBook
		rows.Scan(&boothBook.ID, &boothBook.BoothID, &boothBook.UserID, &boothBook.Period)
		boothBooks = append(boothBooks, boothBook)
	}

	c.JSON(200, gin.H{"booth_books": boothBooks})
}

func GetUserBooksRouter(c *gin.Context) {
	userID := c.Param("user_id")

	rows, _ := db.Query("SELECT * FROM booth_books WHERE user_id = ?", userID)

	var boothBooks []BoothBook
	for rows.Next() {
		var boothBook BoothBook
		rows.Scan(&boothBook.ID, &boothBook.BoothID, &boothBook.UserID, &boothBook.Period)
		boothBooks = append(boothBooks, boothBook)
	}

	c.JSON(200, gin.H{"booth_books": boothBooks})
}

func DeleteBoothBookRouter(c *gin.Context) {
	boothID := c.Param("booth_id")
	period := c.Param("period")

	var respData BoothBookUser
	c.BindJSON(&respData)
	userID := respData.userID

	db.Exec("DELETE FROM booth_books WHERE booth_id = ? AND period = ? AND user_id = ?", boothID, period, userID)

	c.JSON(200, gin.H{"id": boothID})
}

func main() {
	SetupDatabase()

	router := GetRouter()
	router.Run(":3001")
}
