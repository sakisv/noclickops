package main

import (
	"maps"
	"slices"
	"strings"

	claws "github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func filter(managedIds map[string]struct{}, foundResources map[string][]common.Resource, ignoredArns map[string]struct{}) map[string]common.FilteredResults {
	unmanagedResources := make(map[string]common.FilteredResults)
	for key, value := range foundResources {
		if len(value) == 0 {
			continue
		}
		entry := unmanagedResources[key]
		for _, el := range value {
			entry.Meta.Found += 1
			_, found := managedIds[el.TerraformID]

			// first check for things to ignore
			if _, found := ignoredArns[el.Arn]; found {
				entry.Meta.Ignored += 1
				continue
			}

			// then we check for managed resources
			if found {
				entry.Meta.Managed += 1
			} else {
				entry.Resources = append(entry.Resources, el)
				entry.Meta.Unmanaged += 1
			}
		}

		// calculate totals for the service
		if entry.Meta.Unmanaged == 0 {
			entry.Meta.PctUnmanaged = 0
		} else {
			entry.Meta.PctUnmanaged = (float32(entry.Meta.Unmanaged) / float32(entry.Meta.Found)) * 100
		}
		unmanagedResources[key] = entry
	}
	return unmanagedResources
}

func getDefaultIgnoreTags(service claws.NoclickopsResourceGroupTaggingAPIService) []string {
	tagKeyResources := service.GetTagKeysWithPrefixes(IGNORED_TAG_KEY_PREFIXES)

	var tagKeys []string
	for _, key := range tagKeyResources {
		tagKeys = append(tagKeys, key.TerraformID)
	}
	return tagKeys
}

func getIgnoredTagResources(s claws.NoclickopsResourceGroupTaggingAPIService, ignoredTags map[string][]string) map[string]struct{} {
	arns := make(map[string]struct{})

	tagKeys := getDefaultIgnoreTags(s)
	for _, k := range tagKeys {
		if _, found := ignoredTags[k]; found {
			continue
		}
		ignoredTags[k] = make([]string, 0)
	}

	it := strings.Join(slices.Collect(maps.Keys(ignoredTags)), ",")
	println("Ignoring resources tagged with these tags:")
	println(it)

	for k, v := range ignoredTags {
		resources := s.GetResourcesWithTags(k, v)
		for _, r := range resources {
			if _, found := arns[r.Arn]; !found {
				arns[r.Arn] = struct{}{}
			}
		}
	}
	return arns
}

func calculateSummary(filteredResults map[string]common.FilteredResults) common.Summary {
	var summary common.Summary

	for _, v := range filteredResults {
		summary.FoundInAWS += v.Meta.Found
		summary.NotFoundInTerraform += v.Meta.Unmanaged
		summary.FoundInTerraform += v.Meta.Managed
		summary.Ignored += v.Meta.Ignored
	}

	summary.PctUnmanaged = (float32(summary.NotFoundInTerraform) / float32(summary.FoundInAWS)) * 100
	return summary
}
