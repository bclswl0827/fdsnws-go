package provider

import (
	"time"
)

type Provider struct{}

func New() Provider { return Provider{} }

func (Provider) Now() time.Time {
	return time.Now().UTC()
}
