package frontend

import (
	"encoding/json"
	"fmt"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"github.com/raspi/torjuja/frontend"
	"github.com/raspi/torjuja/pkg/db/iface"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type Server struct {
	db        iface.AllowAPI
	rtr       *chi.Mux
	sseServer *sse.Server
}

func New(db iface.AllowAPI) (s *Server) {
	s = &Server{
		db: db,
		sseServer: sse.NewServer(&sse.Options{
			RetryInterval: 5,
			Logger:        log.New(os.Stdout, `SSE: `, 0),
			/*ChannelNameFunc: func(request *http.Request) string {
				log.Printf(`channel name %v`, request.URL.Path)
				return request.URL.Path
			},*/
		}),
	}

	apirouter := chi.NewRouter()
	apirouter.Post(`/allow`, s.apiAllow)

	router := chi.NewRouter()
	router.Use(mw.Recoverer)
	router.Use(mw.RequestID)
	router.Use(mw.Logger)
	router.Use(mw.URLFormat)

	router.Get(`/`, s.frontpage)

	router.Mount(`/api/v1`, apirouter)

	// Javascript and CSS
	router.Get(`/assets/{}`, s.assets)

	// SSE for blocked events
	router.Handle(`/events/blocked`, s.sseServer)

	s.rtr = router

	return s
}

func (srv Server) assets(writer http.ResponseWriter, request *http.Request) {
	if request.URL.RawQuery != `` {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	ext := strings.TrimLeft(path.Ext(path.Base(request.URL.Path)), `.`)

	switch ext {
	case `js`:
		writer.Header().Add(`Content-Type`, `text/javascript`)
	case `css`:
		writer.Header().Add(`Content-Type`, `text/css`)
	case `png`:
		writer.Header().Add(`Content-Type`, `image/png`)
	case `map`:
		writer.Header().Add(`Content-Type`, `application/json`)
	default:
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	fpath := `public/` + strings.TrimLeft(request.URL.Path, `/`)
	f, err := frontend.Assets.Open(fpath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `err: %v`+"\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(writer, f)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `err: %v`+"\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (srv *Server) GetRouter() *chi.Mux {
	return srv.rtr
}

func (srv *Server) frontpage(writer http.ResponseWriter, request *http.Request) {
	if request.URL.RawQuery != `` {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	f, err := frontend.Index.Open(`public/index.html`)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `err: %v`+"\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(writer, f)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `err: %v`+"\n", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (srv Server) getStruct(w io.Writer, iface interface{}) error {
	b, err := json.Marshal(iface)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (srv Server) readStruct(rdr io.ReadCloser, iface interface{}) error {
	defer rdr.Close()
	b, err := io.ReadAll(rdr)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &iface)
	if err != nil {
		return err
	}

	return nil
}

// apiAllow is a Service.httpApi HTTP handler for allowing DNS queries to Service.db that allows DNS query access
func (srv *Server) apiAllow(writer http.ResponseWriter, request *http.Request) {
	var data AllowDTO

	err := srv.readStruct(request.Body, &data)
	if err != nil {
		log.Printf(`error: %v`, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = srv.db.AllowA(data.FQDN)
	if err != nil {
		log.Printf(`error: %v`, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = srv.db.AllowAAAA(data.FQDN)
	if err != nil {
		log.Printf(`error: %v`, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Success
	err = srv.getStruct(writer, ResponseDTO{
		Message: `ok`,
	})
	if err != nil {
		log.Printf(`error: %v`, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (srv *Server) SendMessage(s string, message *sse.Message) {
	srv.sseServer.SendMessage(s, message)
}
