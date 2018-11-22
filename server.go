package mitch

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Server interface {
	Address() net.Addr
	Store() *Store
}

type server struct {
	ctx      context.Context
	address  net.Addr
	listener net.Listener
	opts     serverOpts
	store    *Store
}

type serverOpts struct {
	port int
}

type ServerOpt func(opts *serverOpts)

func WithPort(port int) ServerOpt {
	return func(opts *serverOpts) {
		opts.port = port
	}
}

func NewServer(ctx context.Context, options ...ServerOpt) (Server, error) {
	var opts serverOpts
	for _, o := range options {
		o(&opts)
	}

	s := &server{
		ctx:   ctx,
		opts:  opts,
		store: newStore(),
	}

	err := s.start()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s, nil
}

func (s *server) start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.opts.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.WithStack(err)
	}
	s.listener = listener
	s.address = listener.Addr()

	go func() {
		<-s.ctx.Done()
		listener.Close()
	}()

	go s.serve()
	return nil
}

func (s *server) Address() net.Addr {
	return s.address
}

func (s *server) Store() *Store {
	return s.store
}

type coolHandler func(r *response) error

func (s *server) serve() {
	m := mux.NewRouter()
	handler := func(ch coolHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			res := &response{s: s, w: w, req: req}
			err := func() (retErr error) {
				defer func() {
					if r := recover(); r != nil {
						if rErr, ok := r.(error); ok {
							cause := errors.Cause(rErr)
							if ae, ok := cause.(APIError); ok {
								res.WriteError(ae.status, ae.messages...)
								return
							}
							retErr = rErr
						} else {
							retErr = errors.Errorf("panic: %+v", r)
						}
					}
				}()
				return ch(res)
			}()
			if err != nil {
				res.WriteError(500, fmt.Sprintf("internal error: %+v", err))
			}
		}
	}
	route := func(route string, ch coolHandler) {
		m.HandleFunc(route, handler(ch))
	}
	routePrefix := func(prefix string, ch coolHandler) {
		m.PathPrefix(prefix).Handler(handler(ch))
	}

	route("/profile", func(r *response) error {
		r.CheckAPIKey()
		r.WriteJSON(Any{
			"user": FormatUser(r.currentUser),
		})
		return nil
	})
	routePrefix("/", func(r *response) error {
		r.WriteError(404, "invalid api endpoint")
		return nil
	})

	http.Serve(s.listener, m)
}
