package main

import _ "ariga.io/atlas-provider-gorm/gormschema"

import (
	"downhill-api/graph"
	"log"
	"net/http"
	"os"
	"fmt"
	"context"

	"github.com/joho/godotenv"
	"downhill-api/database"
	"downhill-api/cmd/api"

	"github.com/redis/go-redis/v9"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
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


	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", 
		DB:       0,  
		Protocol: 2,
	})

	ctx := context.Background()

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

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
