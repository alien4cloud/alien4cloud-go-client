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
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	"github.com/pkg/errors"
)

// Command arguments
var url, user, password, appName, propName string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
	flag.StringVar(&propName, "property", "", "Name of the input property to set")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameter
	if appName == "" {
		log.Panic("Mandatory argument 'app' missing (Name of the application to delete)")
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

	envID, err := client.ApplicationService().GetEnvironmentIDbyName(ctx, appName, alien4cloud.DefaultEnvironmentName)
	if err != nil {
		log.Panic(err)
	}

	// Get input propoerties
	topology, err := client.TopologyService().GetTopology(ctx, appName, envID)
	if err != nil {
		log.Panic(err)
	}

	if propName != "" {
		inputProperties := topology.Data.Topology.Inputs

		propDef, ok := inputProperties[propName]
		if !ok {
			log.Panicf("No such input property %s defined in application", propName)
		}

		fmt.Printf("Input property %s: type %+s\n", propName, propDef.Type)

		complexType := false
		dataTypes := topology.Data.DataTypes
		if dataTypes != nil {
			_, complexType = dataTypes[propDef.Type]
		}
		if !complexType {
			// Simple type, just checking if it is a map or list to print the type of elements
			if propDef.Type == "map" || propDef.Type == "list" {
				fmt.Printf("%s of %s\n", propDef.Type, propDef.EntrySchema.Type)
			}
			return
		}

		descRequest := alien4cloud.ComplexToscaTypeDescriptorRequest{
			Dependencies:       topology.Data.Topology.Dependencies,
			PropertyDefinition: propDef,
		}
		description, err := client.CatalogService().GetComplexTOSCAType(ctx, descRequest)
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("Property %s description:\n", propName)
		err = printDescription(description["data"], "    ")
		if err != nil {
			log.Panic(err)
		}
	}

}

func printDescription(desc interface{}, indentation string) error {

	val, ok := desc.(map[string]interface{})
	if !ok || val == nil {
		fmt.Printf("%sDONE\n", indentation)
		return nil
	}

	rawType, ok := val[alien4cloud.TYPE_DESCRIPTION_TYPE_KEY]
	if !ok {
		return errors.Errorf("Found no %s for %+v ", alien4cloud.TYPE_DESCRIPTION_TYPE_KEY, val)

	}
	valType, ok := rawType.(string)
	if !ok {
		return errors.Errorf("Unexpected value for %s: %+v ", alien4cloud.TYPE_DESCRIPTION_TYPE_KEY, rawType)
	}

	switch valType {
	case alien4cloud.TYPE_DESCRIPTION_COMPLEX_TYPE:
		rawProp, ok := val[alien4cloud.TYPE_DESCRIPTION_PROPERTY_TYPE_KEY]
		if !ok {
			return errors.Errorf("Found no %s for %+v ", alien4cloud.TYPE_DESCRIPTION_PROPERTY_TYPE_KEY, val)
		}
		propType, ok := rawProp.(map[string]interface{})
		if !ok {
			return errors.Errorf("Unexpected value for property types: %+v ", rawProp)
		}

		for k, v := range propType {
			fmt.Printf("%sproperty %s:\n", indentation, k)
			err := printDescription(v, indentation+"    ")
			if err != nil {
				return err
			}
		}
	case alien4cloud.TYPE_DESCRIPTION_TOSCA_TYPE:
		rawDef, ok := val[alien4cloud.TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY]
		if !ok {
			return errors.Errorf("Found no %s for %+v ", alien4cloud.TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY, val)
		}
		def, ok := rawDef.(map[string]interface{})
		if !ok {
			return errors.Errorf("Unexpected value for definition: %+v ", rawDef)
		}
		rawType, ok := def["type"]
		if !ok {
			return errors.Errorf("Found no type in definition: %+v ", def)
		}
		defType, ok := rawType.(string)
		if !ok {
			return errors.Errorf("Definition type is not a string: %+v ", rawType)
		}
		required := false
		rawRequired, ok := def["required"]
		if ok {
			required, _ = rawRequired.(bool)
		}
		fmt.Printf("%stype: %s\n", indentation, defType)
		fmt.Printf("%srequired: %t\n", indentation, required)

	case alien4cloud.TYPE_DESCRIPTION_ARRAY_TYPE:
		rawContent, ok := val[alien4cloud.TYPE_DESCRIPTION_CONTENT_TYPE_KEY]
		if !ok {
			return errors.Errorf("Found no %s for %+v ", alien4cloud.TYPE_DESCRIPTION_CONTENT_TYPE_KEY, val)
		}
		fmt.Printf("%stype: array of\n", indentation)
		err := printDescription(rawContent, indentation+"    ")
		if err != nil {
			return err
		}
	case alien4cloud.TYPE_DESCRIPTION_MAP_TYPE:
		rawContent, ok := val[alien4cloud.TYPE_DESCRIPTION_CONTENT_TYPE_KEY]
		if !ok {
			return errors.Errorf("Found no %s for %+v ", alien4cloud.TYPE_DESCRIPTION_CONTENT_TYPE_KEY, val)
		}
		fmt.Printf("%stype: map of\n", indentation)
		err := printDescription(rawContent, indentation+"    ")
		if err != nil {
			return err
		}
	default:
		fmt.Printf("%sunknown type %s\n", indentation, valType)
	}

	return nil
}
