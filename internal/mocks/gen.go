package mocks

import (
	"github.com/lucassabreu/clockify-cli/api"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/util/defaults"
	"github.com/lucassabreu/clockify-cli/pkg/cmdutil"
)

//go:generate mockery --name=Factory --inpackage --with-expecter
type Factory interface {
	cmdutil.Factory
}

//go:generate mockery --name=Config --inpackage --with-expecter
type Config interface {
	cmdutil.Config
}

//go:generate mockery --name=Client --inpackage --with-expecter
type Client interface {
	api.Client
}

//go:generate mockery --name=TimeEntryDefaults --inpackage --with-expecter
type TimeEntryDefaults interface {
	defaults.TimeEntryDefaults
}
