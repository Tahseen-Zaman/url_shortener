package routes

import (
	"github.com/Tahseen-Zaman/url_shortener/database"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(c *fiber.Ctx) error{
	url := c.Params("url")
	r, err := database.CreateClient(0)
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"error": "DB Error",
		})
	}
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err != redis.Nil {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"error": "DB Error",
		})
	}else if err != nil{
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "cannot connect to DB",
		})
	}
	rInr, _ := database.CreateClient(1)
	defer rInr.Close()
	_ = rInr.Incr(database.Ctx, "counter")

	return c.Redirect(value, 301)

}