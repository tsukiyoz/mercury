//go:build wireinject

package follow

import (
	"github.com/google/wire"
	"github.com/lazywoo/mercury/pkg/wego"
)

func InitAPP() *wego.App {
	//wire.Build()
	wire.Build(
		wire.Struct(new(wego.App)),
	)
	return new(wego.App)
}
