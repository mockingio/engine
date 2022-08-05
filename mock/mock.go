package mock

import (
	"io/ioutil"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Mock struct {
	ID      string   `yaml:"id,omitempty" json:"id,omitempty"`
	Name    string   `yaml:"name,omitempty" json:"name,omitempty"`
	Port    string   `yaml:"port,omitempty" json:"port,omitempty"`
	Routes  []*Route `yaml:"routes,omitempty" json:"routes,omitempty"`
	Proxy   *Proxy   `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	options mockOptions
}

func New(opts ...Option) *Mock {
	m := &Mock{
		options: mockOptions{},
	}

	for _, opt := range opts {
		opt(&m.options)
	}

	return m
}

func FromFile(file string, opts ...Option) (*Mock, error) {
	// TODO: Detects file type, support JSON
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "read mock file")
	}

	return FromYaml(string(data), opts...)
}

func FromYaml(text string, opts ...Option) (*Mock, error) {
	decoder := yaml.NewDecoder(strings.NewReader(text))
	m := New(opts...)
	if err := decoder.Decode(m); err != nil {
		return nil, errors.Wrap(err, "decode yaml to mock")
	}
	defaultValues(m)
	if m.options.idGeneration {
		addIDs(m)
	}

	return m, nil
}

func (c Mock) Validate() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Routes, validation.Required),
	)
}

func (c Mock) ProxyEnabled() bool {
	return c.Proxy != nil && c.Proxy.Enabled
}

func defaultValues(m *Mock) {
	for _, r := range m.Routes {
		if r.Method == "" {
			r.Method = http.MethodGet
		}
		for i, res := range r.Responses {
			if res.Status == 0 {
				res.Status = 200
			}
			r.Responses[i] = res
		}
	}
}

// AddIDs Add ids for mock and routes, responses and rules
func addIDs(m *Mock) {
	if m.ID == "" {
		m.ID = newID()
	}
	for _, r := range m.Routes {
		if r.ID == "" {
			r.ID = newID()
		}

		for i, res := range r.Responses {
			if res.ID == "" {
				res.ID = newID()
				r.Responses[i] = res
			}

			for j, rule := range res.Rules {
				if rule.ID == "" {
					rule.ID = newID()
					res.Rules[j] = rule
				}
			}
		}
	}
}

func newID() string {
	return uuid.NewString()
}
