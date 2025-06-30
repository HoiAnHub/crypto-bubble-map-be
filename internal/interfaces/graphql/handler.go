package graphql

import (
	"net/http"

	"crypto-bubble-map-be/graph"
	"crypto-bubble-map-be/internal/infrastructure/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// Handler represents the GraphQL handler
type Handler struct {
	resolver *graph.Resolver
	logger   *logger.Logger
}

// NewHandler creates a new GraphQL handler
func NewHandler(resolver *graph.Resolver, logger *logger.Logger) *Handler {
	return &Handler{
		resolver: resolver,
		logger:   logger,
	}
}

// GraphQLHandler returns a Gin handler for GraphQL requests
func (h *Handler) GraphQLHandler() gin.HandlerFunc {
	// Create GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: h.resolver,
	}))

	return gin.WrapH(srv)
}

// PlaygroundHandler returns a Gin handler for GraphQL playground
func (h *Handler) PlaygroundHandler() gin.HandlerFunc {
	playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
	return gin.WrapH(playgroundHandler)
}

// HealthHandler returns a simple health check for GraphQL
func (h *Handler) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "graphql",
		})
	}
}
