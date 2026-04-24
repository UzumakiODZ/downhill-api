package main

import (
	"context"
	"downhill-api/database"
	"downhill-api/graph"
	"fmt"
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

	const defaultPort = "8080"

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	rdb := redis.NewClient(&redis.Options{})
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)  
		if err != nil {
			log.Printf("Invalid REDIS_URL: %v", err)
		}
		rdb = redis.NewClient(opt)
	}


	ctx := context.Background()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{RedisClient: rdb},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.FixedComplexityLimit(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = rdb.Set(ctx, "foo", "bar", 0).Err()

	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("foo", val)

	database.Connect()
	if database.DB == nil {
		log.Fatal("Database connection failed")
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	rdb.Close()

}
