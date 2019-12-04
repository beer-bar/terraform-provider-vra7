package vra7

import (
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-vra7/utils"
)

// ResourceConfigurationStruct - structure representing the resource_configuration
type ResourceConfigurationStruct struct {
	Name          string                 `json:"name"`
	Configuration map[string]interface{} `json:"configuration"`
}

func resourceConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"configuration": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func expandResourceConfiguration(rConfigurations []interface{}) []ResourceConfigurationStruct {
	configs := make([]ResourceConfigurationStruct, 0, len(rConfigurations))

	for _, config := range rConfigurations {
		configMap := config.(map[string]interface{})

		rConfig := ResourceConfigurationStruct{
			Name:          configMap["name"].(string),
			Configuration: configMap["configuration"].(map[string]interface{}),
		}
		configs = append(configs, rConfig)
	}
	return configs
}

func flattenResourceConfigurations(configs []ResourceConfigurationStruct) []map[string]interface{} {
	if len(configs) == 0 {
		return make([]map[string]interface{}, 0)
	}
	rConfigs := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		componentName, resourceDataMap := parseDataMap(config.Configuration)
		helper := make(map[string]interface{})
		helper["name"] = componentName
		helper["configuration"] = resourceDataMap
		rConfigs = append(rConfigs, helper)
	}
	return rConfigs
}

func parseDataMap(resourceData map[string]interface{}) (string, map[string]interface{}) {
	m := make(map[string]interface{})
	componentName := ""
	resourcePropertyMapper := utils.ResourceMapper()
	for key, value := range resourceData {

		// Component property is within data of a resource, so fetching it from there and putting it as resource level property
		if key == "Component" {
			componentName = convToString(value)
		}
		if i, ok := resourcePropertyMapper[key]; ok {
			key = i
		}
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Slice:
			parseArray(key, m, value.([]interface{}))
		case reflect.Map:
			parseMap(key, m, value.(map[string]interface{}))
		default:
			m[key] = convToString(value)
		}
	}
	return componentName, m
}

func parseMap(prefix string, m map[string]interface{}, data map[string]interface{}) {

	for key, value := range data {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Slice:
			parseArray(prefix+"."+key, m, value.([]interface{}))
		case reflect.Map:
			parseMap(prefix+"."+key, m, value.(map[string]interface{}))
		default:
			m[prefix+"."+key] = convToString(value)
		}
	}
}

func parseArray(prefix string, m map[string]interface{}, value []interface{}) {

	for index, val := range value {
		v := reflect.ValueOf(val)
		switch v.Kind() {
		case reflect.Map:
			/* for properties like NETWORK_LIST, DISK_VOLUMES etc, the value is a slice of map as follows.
			Out of all the information, only data is important information, so leaving out rest of the properties
			 "NETWORK_LIST":[
					{
						"componentTypeId":"",
						"componentId":null,
						"classId":"",
						"typeFilter":null,
						"data":{
						   "NETWORK_MAC_ADDRESS":"00:50:56:b6:78:c6",
						   "NETWORK_NAME":"dvPortGroup-wdc-sdm-vm-1521"
						}
					 }
				  ]
			*/
			objMap := val.(map[string]interface{})
			for k, v := range objMap {
				if k == "data" {
					parseMap(prefix+"."+convToString(index), m, v.(map[string]interface{}))
				}
			}
		default:
			m[prefix+"."+convToString(index)] = convToString(val)
		}
	}
}

func convToString(value interface{}) string {

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return value.(string)
	case reflect.Float64:
		return strconv.FormatFloat(value.(float64), 'f', 0, 64)
	case reflect.Float32:
		return strconv.FormatFloat(value.(float64), 'f', 0, 32)
	case reflect.Int:
		return strconv.Itoa(value.(int))
	case reflect.Int32:
		return strconv.Itoa(value.(int))
	case reflect.Int64:
		return strconv.FormatInt(value.(int64), 10)
	case reflect.Bool:
		return strconv.FormatBool(value.(bool))
	}
	return ""
}
