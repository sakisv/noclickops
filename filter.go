package main

import "github.com/noclickops/common"

func filter(managedIds map[string]struct{}, foundRecords map[string][]common.Resource) map[string][]common.Resource {
	unmanagedResourceIds := make(map[string][]common.Resource)
	for key, value := range foundRecords {
		if len(value) == 0 {
			continue
		}
		for _, el := range value {
			_, found := managedIds[el.TerraformID]
			if found {
				//println("[DEBUG] Found " + el)
			} else {
				unmanagedResourceIds[key] = append(unmanagedResourceIds[key], el)
				//println("[DEBUG] Not found " + el)
			}
		}
	}
	return unmanagedResourceIds
}
