package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/tcarreira/roaw2020/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
