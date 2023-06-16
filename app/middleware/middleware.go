package middleware

import (
	"app/data"

	"github.com/wtran29/fenix/fenix"
)

type Middleware struct {
	App    *fenix.Fenix
	Models data.Models
}
