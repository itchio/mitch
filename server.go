package mitch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

type coolHandler func(r *response)

func (s *server) serve() {
	m := mux.NewRouter()
	handler := func(ch coolHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			res := &response{
				s:     s,
				w:     w,
				req:   req,
				store: s.store,
			}
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
				ch(res)
				return nil
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

	route("/profile", func(r *response) {
		r.RespondTo(RespondToMap{
			"GET": func() {
				r.CheckAPIKey()
				r.WriteJSON(Any{
					"user": FormatUser(r.currentUser),
				})
			},
		})
	})

	route("/games/{id}", func(r *response) {
		r.RespondTo(RespondToMap{
			"GET": func() {
				r.CheckAPIKey()
				gameID := r.Int64Var("id")
				game := r.store.FindGame(gameID)
				r.AssertAuthorization(game.CanBeViewedBy(r.currentUser))
				r.WriteJSON(Any{
					"game": FormatGame(game),
				})
			},
		})
	})

	route("/games/{id}/uploads", func(r *response) {
		r.RespondTo(RespondToMap{
			"GET": func() {
				r.CheckAPIKey()
				gameID := r.Int64Var("id")
				game := r.store.FindGame(gameID)
				r.AssertAuthorization(game.CanBeViewedBy(r.currentUser))
				uploads := r.store.ListUploadsByGame(gameID)
				r.WriteJSON(Any{
					"uploads": FormatUploads(uploads),
				})
			},
		})
	})

	route("/games/{id}/download-sessions", func(r *response) {
		r.RespondTo(RespondToMap{
			"POST": func() {
				r.CheckAPIKey()
				gameID := r.Int64Var("id")
				game := r.store.FindGame(gameID)
				r.AssertAuthorization(game.CanBeViewedBy(r.currentUser))
				r.WriteJSON(Any{
					"uuid": uuid.New().String(),
				})
			},
		})
	})

	route("/uploads/{id}/download", func(r *response) {
		r.RespondTo(RespondToMap{
			"GET": func() {
				r.CheckAPIKey()
				uploadID := r.Int64Var("id")
				upload := r.store.FindUpload(uploadID)
				r.AssertAuthorization(upload.CanBeDownloadedBy(r.currentUser))
				switch upload.Storage {
				case "hosted":
					r.RedirectTo(s.makeURL("/@cdn%s", upload.CDNPath()))
				default:
					Throw(500, "unsupported storage")
				}
			},
		})
	})

	routePrefix("/@cdn", func(r *response) {
		r.RespondTo(RespondToMap{
			"GET": func() {
				path := r.req.URL.Path
				path = strings.TrimPrefix(path, "/@cdn")
				f := r.store.CDNFiles[path]
				if f == nil {
					Throw(404, "not found")
				} else {
					r.Header().Set("content-type", "application/octet-stream")
					r.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=%q", f.Filename))
					r.status = 200
					r.WriteHeader()
					src := bytes.NewReader(f.Contents)
					io.Copy(r.w, src)
				}
			},
		})
	})

	routePrefix("/", func(r *response) {
		Throw(404, "invalid api endpoint")
	})

	http.Serve(s.listener, m)
}
