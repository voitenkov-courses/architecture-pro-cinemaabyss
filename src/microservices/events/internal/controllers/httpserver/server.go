package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/voitenkov-courses/architecture-pro-cinemaabyss/src/microservices/events/api"
	"github.com/voitenkov-courses/architecture-pro-cinemaabyss/src/microservices/movies"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ api.ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// Проверка работоспособности микросервиса событий
// (GET /api/events/health)
func (Server) GetEventsServiceHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

// Создание события фильма
// (POST /api/events/movie)
func (Server) CreateMovieEvent(c *gin.Context) {
	var m movies.Movie
	if err := json.NewDecoder(c.Request.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemeneted",
	})
}

// Создание события платежа
// (POST /api/events/payment)
func (Server) CreatePaymentEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemeneted",
	})
}

// Создание события пользователя
// (POST /api/events/user)
func (Server) CreateUserEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemeneted",
	})
}
