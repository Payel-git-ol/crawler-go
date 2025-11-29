package search

import (
	"context"
	"fmt"

	"github.com/typesense/typesense-go/v4/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
)

func InitTypesense() *typesense.Client {
	client := typesense.NewClient(
		typesense.WithServer("http://localhost:8108"),
		typesense.WithAPIKey("xyz"),
	)

	schema := &api.CollectionSchema{
		Name: "issues",
		Fields: []api.Field{
			{Name: "id", Type: "int32"},
			{Name: "title", Type: "string"},
			{Name: "url", Type: "string"},
			{Name: "state", Type: "string"},
		},
	}

	_, err := client.Collections().Create(context.Background(), schema)
	if err != nil {
		fmt.Println("Typesense init:", err)
	}

	return client
}
