package main

import "github.com/noclickops/common"

func filter(managedIds map[string]struct{}, foundRecords map[string][]common.Resource) map[string]common.FilteredResults {
	unmanagedResources := make(map[string]common.FilteredResults)
	for key, value := range foundRecords {
		if len(value) == 0 {
			continue
		}
		entry := unmanagedResources[key]
		for _, el := range value {
			entry.Meta.Found += 1
			_, found := managedIds[el.TerraformID]
			if found {
				entry.Meta.Managed += 1
			} else {
				entry.Resources = append(entry.Resources, el)
				entry.Meta.Unmanaged += 1
			}
		}
		unmanagedResources[key] = entry
	}
	return unmanagedResources
}
