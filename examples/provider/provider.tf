terraform {
  required_providers {
    metabase = {
      source  = "bnjns/metabase"
      version = "~> 0.0"
    }
  }
}

provider "metabase" {
  host = "<your metabase host/ip>"
}
