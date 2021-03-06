package models

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	pathToFileCache = cache.New(5*time.Minute, 10*time.Minute)
)
