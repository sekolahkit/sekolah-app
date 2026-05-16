package notifikasi

import (
	"fmt"
)

type SendResult struct {
	Success bool
	Error   error
}

type Notifier interface {
	Send(n *Notifikasi) SendResult
	Type() string
}

type Registry struct {
	providers map[string]Notifier
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Notifier)}
}

func (r *Registry) Register(n Notifier) {
	r.providers[n.Type()] = n
}

func (r *Registry) Get(tipe string) (Notifier, error) {
	n, ok := r.providers[tipe]
	if !ok {
		return nil, fmt.Errorf("provider %q tidak tersedia", tipe)
	}
	return n, nil
}

func (r *Registry) Has(tipe string) bool {
	_, ok := r.providers[tipe]
	return ok
}

func (r *Registry) Types() []string {
	var t []string
	for k := range r.providers {
		t = append(t, k)
	}
	return t
}
