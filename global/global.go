package global

import (
	"github.com/Yosorable/ms-metadata/core/config"
	"gorm.io/gorm"
)

var (
	CONFIG   *config.Config
	DATABASE *gorm.DB
)
