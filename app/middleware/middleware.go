package middleware

import (
	"app/data"

	"github.com/wtran29/fenix"
)

type Middleware struct {
	App    *fenix.Fenix
	Models data.Models
}
