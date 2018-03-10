package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/olivere/elastic"
	"github.com/satori/go.uuid"
)

const index = "search"

var categories = []string{"restaurant", "bar", "clerb"}

type ElasticStore struct {
	client *elastic.Client
}

func NewElasticStore(client *elastic.Client) *ElasticStore {
	return &ElasticStore{
		client: client,
	}
}

func (es *ElasticStore) Fetch(from int, to int, category string) (*Results, error) {
	ctx := context.Background()
	size := to - from

	var results = &Results{
		Results: []Result{},
		Token:   to,
	}
	// Combine both into a boolquery
	boolQuery := elastic.NewBoolQuery()

	// individual requirements
	termQuery := elastic.NewTermQuery("Category", category)

	boolQuery.Must(termQuery)

	searchResult, err := es.client.Search().
		Index(index).
		Query(boolQuery).
		From(from).Size(size).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	fmt.Printf("found %v", searchResult.Hits.TotalHits)
	var res Result
	for _, item := range searchResult.Each(reflect.TypeOf(res)) {
		r := item.(Result)
		results.Results = append(results.Results, r)
	}

	return results, nil
}

////////////////////////////
//     Initialization     //
////////////////////////////

// The bulk of these setup steps were lifted from https://github.com/olivere/elastic/wiki/Services

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"item":{
			"properties":{
				"ID":{
					"type":"text"
				},
				"Name":{
					"type":"text"
				},
				"Category":{
					"type":"keyword"
				}
			}
		}
	}
}`

// SetupIndex checks for or creates the correct index.
func (es *ElasticStore) SetupIndex() error {
	ctx := context.Background()
	exists, err := es.client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create a new index.
		createIndex, err := es.client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			fmt.Println("Index was created, but not acknowledge. It may still be initializing.")
		}
	}

	return nil
}

// SeedData seeds the index with *count* pieces of data. Since it uses uuids, it'll seed new data whenever it's run.
func (es *ElasticStore) SeedData(count int) error {
	rand.Seed(time.Now().Unix())

	ctx := context.Background()
	bulkRequest := es.client.Bulk()

	for i := 0; i < count; i++ {
		id := uuid.NewV4().String()

		item := Result{
			ID:       id,
			Name:     fmt.Sprintf("Super Neat Place %v", i),
			Category: categories[rand.Intn(len(categories))],
		}

		indexRequest := elastic.NewBulkIndexRequest().Index(index).Type("item").Id(id).Doc(item)
		bulkRequest = bulkRequest.Add(indexRequest)
	}

	bulkResponse, err := bulkRequest.Do(ctx)
	if err != nil {
		return err
	}

	if bulkRequest.NumberOfActions() != 0 {
		return errors.New("Leftover bulk actions")
	}

	// Indexed returns information abount indexed documents
	indexed := bulkResponse.Indexed()
	if len(indexed) != count {
		return errors.New(fmt.Sprintf("Mismatched number of documents indexed. Expected %v, got %v", count, indexed))
	}

	failedResults := bulkResponse.Failed()
	if len(failedResults) != 0 {
		fmt.Printf("Failed to insert %v/%v documents in bulkInsert", len(failedResults), count)
	}
	return nil
}
