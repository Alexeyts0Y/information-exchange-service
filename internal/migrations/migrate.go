package migrations

import "embed"

// FS — встроенные SQL-миграции. Позволяет поставлять один бинарь без внешних файлов.
//
//go:embed *.sql
var FS embed.FS
