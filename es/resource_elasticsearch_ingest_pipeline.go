package es

import (
	"context"
	"fmt"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

func resourceElasticsearchIngestPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticsearchIngestPipelineCreate,
		Read:   resourceElasticsearchIngestPipelineRead,
		Update: resourceElasticsearchIngestPipelineUpdate,
		Delete: resourceElasticsearchIngestPipelineDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"body": {
				Type:             schema.TypeString,
				DiffSuppressFunc: diffSuppressIngestPipeline,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceElasticsearchIngestPipelineCreate(d *schema.ResourceData, meta interface{}) error {

	err := resourceElasticsearchPutIngestPipeline(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticsearchIngestPipelineRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	client := meta.(*elastic.Client)
	res, err := client.Ingest.GetPipeline(
		client.Ingest.GetPipeline.WithPipelineID(id),
		client.Ingest.GetPipeline.WithContext(context.Background()),
		client.Ingest.GetPipeline.WithPretty(),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Ingest Pipeline %s not found - removing from state", id)
			log.Warnf("Ingest Pipeline %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when getting Ingest pipeline %s: %s", id, res.String())

	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	body := string(b)

	log.Debugf("Got ingest pipeline %s successfully:\n%s", id, body)
	d.Set("name", d.Id())
	d.Set("body", body)
	return nil

}

func resourceElasticsearchIngestPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticsearchPutIngestPipeline(d, meta)
}

func resourceElasticsearchIngestPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	client := meta.(*elastic.Client)

	res, err := client.Ingest.DeletePipeline(
		id,
		client.Ingest.DeletePipeline.WithContext(context.Background()),
		client.Ingest.DeletePipeline.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			fmt.Printf("[WARN] Ingest Pipeline %s not found - removing from state", id)
			log.Warnf("Ingest Pipeline %s not found - removing from state", id)
			d.SetId("")
			return nil
		}
		return errors.Errorf("Error when delete ingest pipeline %s: %s", id, res.String())

	}

	d.SetId("")
	return nil
}

func resourceElasticsearchPutIngestPipeline(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	client := meta.(*elastic.Client)

	res, err := client.Ingest.PutPipeline(
		name,
		strings.NewReader(body),
		client.Ingest.PutPipeline.WithContext(context.Background()),
		client.Ingest.PutPipeline.WithPretty(),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error when putting ingest pipeline %s: %s", name, res.String())
	}

	return nil
}