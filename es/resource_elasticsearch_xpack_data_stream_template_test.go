package es

import (
	"context"
	"fmt"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchDataStream(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchDataStreamTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchDataStreamTemplate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchDataStreamTemplateExists("elasticsearch_xpack_data_stream_template.test"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testElasticsearchDataStreamTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchDataStreamTemplateExists("elasticsearch_xpack_data_stream_template.test"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      "elasticsearch_xpack_data_stream_template.test",
				ImportState:       true,
				ImportStateVerify: true,
				//ImportStateVerifyIgnore: []string{"index_templates", "data_stream", "template", "priority"},
			},
		},
	})
}

func testCheckElasticsearchDataStreamTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No data stream ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Indices.GetIndexTemplate(
			client.API.Indices.GetIndexTemplate.WithName(rs.Primary.ID),
			client.API.Indices.GetIndexTemplate.WithContext(context.Background()),
			client.API.Indices.GetIndexTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get data stream template %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchDataStreamTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_xpack_data_stream_template" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Indices.DeleteIndexTemplate(
			rs.Primary.ID,
			client.API.Indices.DeleteIndexTemplate.WithContext(context.Background()),
			client.API.Indices.DeleteIndexTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				return nil
			}
		}

		return fmt.Errorf("Data stream template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchDataStreamTemplate = `
resource "elasticsearch_xpack_data_stream_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test*"
  ],
  "data_stream": {},
  "template": {
    "settings": {
      "index.lifecycle.name": "my-data-stream-policy"
    }
  },
  "priority": 20
}
EOF
}
`

var testElasticsearchDataStreamTemplateUpdate = `
resource "elasticsearch_xpack_data_stream_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test*"
  ],
  "data_stream": {},
  "template": {
    "settings": {
      "index.lifecycle.name": "my-data-stream-policy-2"
    }
  },
  "priority": 23
}
EOF
}
`
