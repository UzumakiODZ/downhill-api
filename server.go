package main

import (
	"context"
	"downhill-api/database"
	"downhill-api/graph"
	"log"
	"net/http"
	"os"

	_ "ariga.io/atlas-provider-gorm/gormschema"
	"github.com/joho/godotenv"

	cmd "downhill-api/cmd/api"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/redis/go-redis/v9"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/go-chi/chi/v5"  
    "github.com/rs/cors"
)

func graphqlHandler(srv *handler.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if token, ok := r.Context().Value("authToken").(string); ok {
				cmd.SetAuthCookie(w, token)
			}
			if r.Context().Value("logout") == true {
				cmd.ClearAuthCookie(w)
			}
		}()
		srv.ServeHTTP(w, r)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file")
	}

	const defaultPort = "8080"
	const defaultRedisURL = "redis://127.0.0.1:6379/0"

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = defaultRedisURL
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Invalid REDIS_URL: %v. Falling back to %s", err, defaultRedisURL)
		opt, _ = redis.ParseURL(defaultRedisURL)
	}
	rdb := redis.NewClient(opt)
	defer func() {
		_ = rdb.Close()
	}()

	ctx := context.Background()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{RedisClient: rdb},
	}))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://downhill-lovat.vercel.app",
			"https://downhill-lovat.vercel.app/",  
			"http://localhost:5173",
		},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept", 
			"Content-Type",           
			"Authorization",
			"sec-ch-ua",              
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		},
		AllowCredentials: true,
	}).Handler

	router := chi.NewRouter()
    router.Use(corsHandler)  

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.FixedComplexityLimit(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis not reachable (%v). Server will continue, but cache/session features may fail.", err)
	} else {
		log.Printf("Redis connected")
	}

	database.Connect()
	if database.DB == nil {
		log.Fatal("Database connection failed")
	}

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}
