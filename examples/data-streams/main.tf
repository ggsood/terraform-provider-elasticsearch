terraform {
  required_version = ">= 0.12.29"

  required_providers {
    elasticsearch = {
      source  = "ggsood/elasticsearch"
      version = "0.1.0-beta"
    }
  }
}

provider "elasticsearch" {
  hosts = "https://38b970e313304d3bb3bf3da116aa053b.eu-west-2.aws.cloud.es.io:9243"
  username = "elastic"
  password = "P4N4kuipuqdpN2WMXx6GBvGc"
}

resource elasticsearch_data_stream_template "test" {
  name         = "terraform-data-stream-test"
  template     = <<EOF
{
  "index_patterns": [
    "test"
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