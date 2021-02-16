package es

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// resourceElasticsearchDataStreamTemplate handle the index template API call
func resourceElasticsearchDataStreamTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchDataStreamTemplateCreate,
		Update: resourceElasticsearchDataStreamTemplateUpdate,
		Read:   resourceElasticsearchDataStreamTemplateRead,
		Delete: resourceElasticsearchDataStreamTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"template": {
				Type:             schema.TypeString,
				Required:         true,
			},
		},
	}
}

// resourceElasticsearchDataStreamTemplateCreate create index template
func resourceElasticsearchDataStreamTemplateCreate(d *schema.ResourceData, meta interface{}) error {

	err := createDataStreamTemplate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return resourceElasticsearchDataStreamTemplateRead(d, meta)
}

// resourceElasticsearchDataStreamTemplateUpdate update index template
func resourceElasticsearchDataStreamTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	err := createDataStreamTemplate(d, meta)
	if err != nil {
		return err
	}
	return resourceElasticsearchDataStreamTemplateRead(d, meta)
}

// resourceElasticsearchDataStreamTemplateRead read index template
func resourceElasticsearchDataStreamTemplateRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.GetIndexTemplate(
		client.API.Indices.GetIndexTemplate.WithName(id),
		client.API.Indices.GetIndexTemplate.WithContext(context.Background()),
		client.API.Indices.GetIndexTemplate.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Data Stream template %s not found - removing from state", id)
			log.Warnf("Data Stream template %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when get Data Stream template %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	body := string(b)

	log.Debugf("Get data stream template %s successfully:\n%s", id, body)
	d.Set("name", d.Id())
	d.Set("template", body)
	return nil
}

// resourceElasticsearchDataStreamTemplateDelete delete index template
func resourceElasticsearchDataStreamTemplateDelete(d *schema.ResourceData, meta interface{}) error {

	id := d.Id()

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.DeleteIndexTemplate(
		id,
		client.API.Indices.DeleteIndexTemplate.WithContext(context.Background()),
		client.API.Indices.DeleteIndexTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Data Stream template %s not found - removing from state", id)
			log.Warnf("Data Stream template %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete index template %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

// createDataStreamTemplate create or update data stream template
func createDataStreamTemplate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	template := d.Get("template").(string)

	client := meta.(*elastic.Client)
	res, err := client.API.Indices.PutIndexTemplate(
		name,
		strings.NewReader(template),
		client.API.Indices.PutIndexTemplate.WithContext(context.Background()),
		client.API.Indices.PutIndexTemplate.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when add index template %s: %s", name, res.String())
	}

	return nil
}
