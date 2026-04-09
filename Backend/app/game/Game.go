package game

import(
	"github.com/gin-gonic/gin"
)

func ListGames(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var games []Game
		db.Select(&games, "SELECT * FROM games")
		c.JSON(200, games)
	}
}
