package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	//"net/url"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
)

type Client struct {
	baseURL string
}

// Get returns the PKI document for the provided epoch.
func (c *Client) Get(ctx context.Context, epoch uint64) (*pki.Document, error) {
	url := fmt.Sprintf("%s/consensus/%d", c.baseURL, epoch)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	r := request.WithContext(ctx)
	httpClient := http.Client{}
	response, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("Client Get failure: status not 200")
	}
	var buf bytes.Buffer
	buf.ReadFrom(response.Body)
	document := pki.Document{}
	err = json.Unmarshal(buf.Bytes(), &document)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// Post posts the node's descriptor to the PKI for the provided epoch.
func (c *Client) Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *pki.MixDescriptor) error {
	descriptor, err := json.Marshal(d)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(descriptor)
	url := fmt.Sprintf("%s/upload/", c.baseURL)
	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}
	r := request.WithContext(ctx)
	httpClient := http.Client{}
	response, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("Client Post fail, status not 200")
	}
	return nil
}

func main() {
	c := Client{
		baseURL: "http://127.0.0.1:8080/B",
	}
	ctx := context.TODO()
	doc, err := c.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println("doc", doc)
}
