package rueidisleader

import (
	_ "embed"

	"github.com/redis/rueidis"
)

var (
	//go:embed scripts/obtain.lua
	obtainRaw string

	//go:embed scripts/release.lua
	releaseRaw string

	//go:embed scripts/renew.lua
	renewRaw string

	//go:embed scripts/check.lua
	checkRaw string

	obtain  = rueidis.NewLuaScript(obtainRaw)
	release = rueidis.NewLuaScript(releaseRaw)
	renew   = rueidis.NewLuaScript(renewRaw)
	check   = rueidis.NewLuaScript(checkRaw)
)
