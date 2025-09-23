package http

import (
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/config"
	"VKR_gateway_service/internal/transport/http/middlewares"
	"context"
	"fmt"
	"net/http"
	"net/url"

	docs "VKR_gateway_service/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	domain     string
	port       string
	app        *gin.Engine
	httpServer *http.Server
}

func NewHTTPServer(conf *config.Config, a *app.App) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gin.Logger(),
	)
	httpServer := &http.Server{
		Addr:    ":" + conf.HttpServerConfig.Port,
		Handler: r,
	}
	s := Server{
		domain:     conf.Domain,
		port:       conf.HttpServerConfig.Port,
		app:        r,
		httpServer: httpServer,
	}

	allowed := conf.AllowedCORSOrigins
	if len(allowed) == 0 {
		a.Logger.Fatalf("no allowed CORS origins configured")
		return nil
	}

	docs.SwaggerInfo.BasePath = "/api"

	if conf.PublicURL != "" {
		if u, err := url.Parse(conf.PublicURL); err == nil {
			docs.SwaggerInfo.Host = u.Host
			if u.Scheme == "https" {
				docs.SwaggerInfo.Schemes = []string{"https"}
			} else {
				docs.SwaggerInfo.Schemes = []string{"http"}
			}
		}
	} else {
		docs.SwaggerInfo.Host = conf.Domain + ":" + conf.HttpServerConfig.Port
		if len(docs.SwaggerInfo.Schemes) == 0 {
			docs.SwaggerInfo.Schemes = []string{"http"}
		}
	}

	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     allowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	if conf.SwaggerEnabled {
		if conf.SwaggerUser != "" && conf.SwaggerPassword != "" {
			authorized := s.app.Group("/swagger", gin.BasicAuth(gin.Accounts{
				conf.SwaggerUser: conf.SwaggerPassword,
			}))
			authorized.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		} else {
			s.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	}
	// Public routers
	SSORouter(s.app.Group("/api/sso/"), a)

	// Protected routers
	ai := s.app.Group("/api/ai/")
	ai.Use(middlewares.AuthMiddleware(a))
	AIRouter(ai, a)
	return &s
}

func (s *Server) Listen() error {
	fmt.Printf("Server is running on %s:%s\n", s.domain, s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
