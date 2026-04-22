package fdsnwsgo

import (
	"fmt"
	"strconv"
	"strings"

	component "github.com/bclswl0827/fdsnws-go/service"
	"github.com/bclswl0827/fdsnws-go/service/dataselect"
	"github.com/bclswl0827/fdsnws-go/service/event"
	"github.com/bclswl0827/fdsnws-go/service/station"
)

type config struct {
	basePath          string
	documentationURL  string
	selected          map[Module]bool
	modulesConfigured bool
	customServices    map[Module]component.Service
	versions          map[Module]string
	authenticator     component.Authenticator
}

type Option func(*config) error

// WithModules enables exactly the listed service names. Custom services can
// be selected by converting their name to Module.
func WithModules(modules ...Module) Option {
	return func(c *config) error {
		c.modulesConfigured = true
		c.selected = make(map[Module]bool, len(modules))
		for _, module := range modules {
			if err := validateServiceName(module); err != nil {
				return err
			}
			c.selected[module] = true
		}
		return nil
	}
}

// WithService registers or replaces a service component. If WithModules is
// not used, registered components are enabled automatically.
func WithService(service Service) Option {
	return func(c *config) error {
		if isNil(service) {
			return fmt.Errorf("service component cannot be nil")
		}
		if err := validateServiceName(service.Name()); err != nil {
			return err
		}
		c.customServices[service.Name()] = service
		return nil
	}
}

func WithBasePath(path string) Option {
	return func(c *config) error {
		path = "/" + strings.Trim(path, "/")
		if path == "/" {
			return fmt.Errorf("base path cannot be empty")
		}
		c.basePath = path
		return nil
	}
}

func WithDocumentationURL(uri string) Option {
	return func(c *config) error { c.documentationURL = uri; return nil }
}

// WithAuthenticator enables the optional dataselect queryauth endpoint. The
// authenticator receives normalized credential input and is independent of
// the data Provider and HTTP server framework.
func WithAuthenticator(authenticator Authenticator) Option {
	return func(c *config) error {
		if isNil(authenticator) {
			return fmt.Errorf("authenticator cannot be nil")
		}
		c.authenticator = authenticator
		return nil
	}
}

func WithServiceVersion(module Module, version string) Option {
	return func(c *config) error {
		if err := validateServiceName(module); err != nil {
			return err
		}
		if _, err := parseVersion(version); err != nil {
			return err
		}
		c.versions[module] = version
		return nil
	}
}

// NewFDSNWS constructs the service registry. Provider may be nil only when no
// enabled built-in component needs it (for example a custom-only registry).
func NewFDSNWS(provider IProvider, options ...Option) (*FDSNWS, error) {
	c := config{
		basePath:       defaultBasePath,
		selected:       make(map[Module]bool),
		customServices: make(map[Module]component.Service),
		versions:       make(map[Module]string),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&c); err != nil {
			return nil, err
		}
	}

	services := make(map[Module]component.Service, len(c.customServices)+3)
	for name, custom := range c.customServices {
		services[name] = custom
	}
	builtIns := []struct {
		name  Module
		build func(IProvider) Service
	}{
		{ModuleDataSelect, func(provider IProvider) Service { return dataselect.New(provider, c.authenticator) }},
		{ModuleStation, func(provider IProvider) Service { return station.New(provider) }},
		{ModuleEvent, func(provider IProvider) Service { return event.New(provider) }},
	}
	for _, builtIn := range builtIns {
		name := builtIn.name
		if _, replaced := services[name]; replaced {
			continue
		}
		if c.modulesConfigured && !c.selected[name] {
			continue
		}
		if isNil(provider) {
			return nil, fmt.Errorf("provider cannot be nil while built-in service %q is enabled", name)
		}
		services[name] = builtIn.build(provider)
	}
	if c.modulesConfigured {
		for name := range services {
			if !c.selected[name] {
				delete(services, name)
			}
		}
		for name := range c.selected {
			if _, ok := services[name]; !ok {
				return nil, fmt.Errorf("service %q is enabled but not registered", name)
			}
		}
	}
	for name, service := range services {
		if service.Name() != name {
			return nil, fmt.Errorf("registered service name changed from %q to %q", name, service.Name())
		}
		majorValue, err := strconv.Atoi(service.MajorVersion())
		if err != nil || majorValue < 1 {
			return nil, fmt.Errorf("service %q has an invalid major version", name)
		}
		version := service.Version()
		if override, ok := c.versions[name]; ok {
			version = override
		}
		major, err := parseVersion(version)
		if err != nil {
			return nil, fmt.Errorf("service %q: %w", name, err)
		}
		if major != service.MajorVersion() {
			return nil, fmt.Errorf("service %q implementation version and major route version differ", name)
		}
	}
	var rootProvider component.IRootProvider
	var timeProvider component.ITimeProvider = component.UnimplementedProvider{}
	if !isNil(provider) {
		rootProvider = provider
		timeProvider = provider
	}
	return &FDSNWS{services: services, rootProvider: rootProvider, timeProvider: timeProvider, config: c}, nil
}

func New(provider IProvider, options ...Option) (*FDSNWS, error) {
	return NewFDSNWS(provider, options...)
}

func parseVersion(version string) (string, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("service version must have three numeric components")
	}
	for _, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return "", fmt.Errorf("service version components must be non-negative integers")
		}
	}
	return parts[0], nil
}

func validateServiceName(module Module) error {
	name := string(module)
	if name == "" || strings.ContainsAny(name, "/?#") {
		return fmt.Errorf("invalid service name %q", name)
	}
	return nil
}
