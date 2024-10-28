package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/Tahseen-Zaman/url_shortener/database"
	"github.com/Tahseen-Zaman/url_shortener/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}
type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset int           `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	err := c.BodyParser(body)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "cannot parse json",
		})
	}

	// implement the rate limiting
	r2, _ := database.CreateClient(1)
	defer r2.Close()
	value, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()

	} else {
		// value, _ = r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(value)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			c.Status(fiber.StatusNotFound)
			return c.JSON(fiber.Map{
				"error": "Rate Limit Exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// check if the input is a actual url
	if !govalidator.IsURL(body.URL) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}
	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}
	// enforce https, ssl
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[0:6]
	} else {
		id = body.CustomShort
	}
	r, _ := database.CreateClient(0)
	defer r.Close()
	value, _ = r.Get(database.Ctx, id).Result()
	if value != "" {
		c.Status(fiber.StatusForbidden)
		return c.JSON(fiber.Map{
			"error": "URL custom short is already in use",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*time.Second).Err()
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	r2.Decr(database.Ctx, c.IP())

	value, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(value)
	limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = int(limit / time.Nanosecond / time.Minute)

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)

}
