package assets

import "embed"

//go:embed index.ghtm
//go:embed iplist.ghtm
var Assets embed.FS
