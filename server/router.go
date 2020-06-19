package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func CORS(origins []string) func(http.Handler) http.Handler {
	defaultCORS := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	corsMapPool := &sync.Map{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := r.Header.Get("Access-Control-Request-Headers")
			if headers != "" {
				var p *sync.Pool
				if v, ok := corsMapPool.Load(headers); ok {
					p = v.(*sync.Pool)
				} else {
					opt := cors.Options{
						AllowedOrigins:   origins,
						AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
						AllowedHeaders:   strings.Split(headers, ","),
						AllowCredentials: true,
						MaxAge:           3600,
					}
					for i, header := range opt.AllowedHeaders {
						opt.AllowedHeaders[i] = strings.TrimSpace(header)
					}

					p = &sync.Pool{
						New: func() interface{} {
							return cors.New(opt)
						},
					}
					corsMapPool.Store(headers, p)
				}
				cc := p.Get().(*cors.Cors)
				cc.Handler(next).ServeHTTP(w, r)
				p.Put(cc)
			} else {
				defaultCORS.Handler(next).ServeHTTP(w, r)
			}
		})
	}
}

func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// r.Mount("/", http.FileServer(http.Dir("./")))
	r.With(CORS([]string{"*"})).HandleFunc("/ws", WebSocket)
	r.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"token": "xxxxx"})
	})
	return r
}
