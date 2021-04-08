package frontend

import "embed"

//go:embed public/assets/*.js public/assets/*.css public/assets/*.png
var Assets embed.FS

//go:embed public/index.html
var Index embed.FS
