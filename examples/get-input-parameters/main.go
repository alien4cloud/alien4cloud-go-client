// Copyright 2020 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, appTemplate string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appTemplate, "template", "", "Name of the topology template to use")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameter
	if appTemplate == "" {
		log.Panic("Mandatory argument 'template' missing (Name of the topology template to use)")
	}

	client, err := alien4cloud.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	// Timeout after one hour (this is optional you can use a context without timeout or cancelation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	err = client.Login(ctx)
	if err != nil {
		log.Panic(err)
	}

	// Get input properties
	topology, err := client.TopologyService().GetTopologyByID(ctx, appTemplate)
	if err != nil {
		log.Panic(err)
	}

	inputProperties := topology.Data.Topology.Inputs
	for propName, propDef := range inputProperties {
		fmt.Printf("Input property %s\n", propName)
		fmt.Printf("\ttype: %s\n", propDef.Type)
		if propDef.Required {
			fmt.Println("\tinput required")
		} else {
			fmt.Printf("\tdefault value: %+v\n", propDef.DefaultValue)
		}
		componentPropNames, err := getComponentPropertiesReferencingInput(topology.Data.Topology.NodeTemplates, propName)
		if err != nil {
			log.Panic(err)
		}

		if len(componentPropNames) > 0 {
			fmt.Println("\treferenced in:")
		}

		for compName, propNames := range componentPropNames {
			fmt.Printf("\t- component %s, properties: %v\n", compName, propNames)

		}
	}

}

func getComponentPropertiesReferencingInput(nodeTemplates map[string]alien4cloud.NodeTemplate, propName string) (map[string][]string, error) {
	result := make(map[string][]string)
	var err error

	for compName, nodeTemplate := range nodeTemplates {
		var propNames []string
		for _, prop := range nodeTemplate.Properties {
			if prop.Value.FunctionConcat != "" {
				// Check if the input property is referenced in one of the
				// concat argument
				for _, param := range prop.Value.Parameters {

					var propValue alien4cloud.PropertyValue
					mapValue, ok := param.(map[string]interface{})
					if ok {
						// transform it into a PropertyValue if applicable
						jsonbody, err := json.Marshal(mapValue)
						if err == nil {
							err = json.Unmarshal(jsonbody, &propValue)
						}
						ok = (err == nil)
					}
					if ok && propValue.Function == alien4cloud.FunctionGetInput &&
						isPropertyUsedInPropertyValueParameters(propName, propValue) {

						propNames = append(propNames, prop.Key)
					}
				}
			} else if prop.Value.Function == alien4cloud.FunctionGetInput &&
				isPropertyUsedInPropertyValueParameters(propName, prop.Value) {

				propNames = append(propNames, prop.Key)
			}

		}

		if len(propNames) > 0 {
			result[compName] = propNames
		}
	}
	return result, err
}

func isPropertyUsedInPropertyValueParameters(propName string, propVal alien4cloud.PropertyValue) bool {
	for _, param := range propVal.Parameters {
		paramVal, ok := param.(string)
		if ok && paramVal == propName {
			return true
		}
	}

	return false
}
