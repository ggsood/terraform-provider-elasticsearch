package es

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/helper/pathorcontents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Provider variables required for calling Elasticsearch APIs
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hosts": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_HOSTS", nil),
				Description: "Elasticsearch Hosts",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
				Description: "Username to use to connect to elasticsearch using basic auth",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
				Description: "Password to use to connect to elasticsearch using basic auth",
			},
			"ca_cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Custom CA certificate",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls",
			},
			"retry": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     6,
				Description: "Number of times to retry connection before failing",
			},
			"wait_before_retry": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Wait time in second before retry connection",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"elasticsearch_index_template": resourceElasticsearchDataStreamTemplate(),
			"elasticsearch_data_stream_template": resourceElasticsearchDataStreamTemplate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure to initialize the rest client to access on Elasticsearch API
func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var (
		data map[string]interface{}
	)

	hosts := strings.Split(d.Get("hosts").(string), ",")
	insecure := d.Get("insecure").(bool)
	caCertFile := d.Get("ca_cert_file").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	retry := d.Get("retry").(int)
	waitBeforeRetry := d.Get("wait_before_retry").(int)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	// Checks is valid URLs
	for _, rawURL := range hosts {
		_, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
	}

	// Initialize connection
	cfg := elastic.Config{
		Addresses: hosts,
	}
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}
	if insecure == true {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	// If a caCertFile has been specified, use that for cert validation
	if caCertFile != "" {
		caCert, _, _ := pathorcontents.Read(caCertFile)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(caCert))
		transport.TLSClientConfig.RootCAs = caCertPool
	}
	cfg.Transport = transport
	client, err := elastic.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Test connection and check elastic version to use the right Version
	numFailed := 0
	isOnline := false
	var res *esapi.Response
	for isOnline == false {
		res, err = client.API.Info(
			client.API.Info.WithContext(context.Background()),
		)
		if err == nil && res.IsError() == false {
			isOnline = true
		} else {
			if numFailed == retry {
				return nil, err
			}
			numFailed++
			time.Sleep(time.Duration(waitBeforeRetry) * time.Second)
		}
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.Errorf("Error when getting info about Elasticsearch client: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	version := data["version"].(map[string]interface{})["number"].(string)
	log.Debugf("Server: %s", version)

	if version < "7.0.0" || version >= "8.0.0" {
		return nil, errors.Errorf("ElasticSearch version is not 7.x (%s), you need to use the right version of elasticsearch provider", version)
	}

	return client, nil
}
