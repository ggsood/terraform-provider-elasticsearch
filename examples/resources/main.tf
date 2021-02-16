terraform {
  required_version = ">= 0.12.29"

  required_providers {
    elasticsearch = {
      source  = "ggsood/elasticsearch"
      version = "0.1.0-beta"
    }
  }
}

provider "elasticsearch" {}

resource "elasticsearch_xpack_data_stream_template" "abc" {
  name         = "terraform-data-stream-test"
  template     = <<EOF
{
  "index_patterns": [
    "gaurav"
  ],
  "data_stream": {},
  "template": {
    "settings": {
      "index.lifecycle.name": "my-data-stream-policy"
    },
    "mappings": {
        "dynamic_templates": [
            {
              "integers": {
                "match_mapping_type": "long",
                "mapping": {
                  "type": "integer"
                }
              }
            }
        ],
        "properties": {
          "name" : {
            "type": "keyword"
          }
        }
    }
  },
  "priority": 20
}
EOF
}

resource "elasticsearch_index_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test"
  ],
  "settings": {
    "index.refresh_interval": "5s",
    "index.lifecycle.name": "policy-logstash-backup",
    "index.lifecycle.rollover_alias": "logstash-backup-alias"
  },
  "order": 2
}
EOF
}

resource "elasticsearch_ingest_pipeline" "test" {
  name = "terraform-test"
  body = <<EOF
{
  "description" : "describe pipeline",
  "version": 123,
  "processors" : [
    {
      "set" : {
        "field": "foo",
        "value": "bar"
      }
    },
    {
      "set" : {
        "field": "abc",
        "value": "xyz"
      }
    }
  ]
}
EOF
}