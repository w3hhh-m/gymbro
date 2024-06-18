package main

import (
	"GYMBRO/internal/config"
)

func main() {
	cfg := config.MustLoad()
	_ = cfg
}
