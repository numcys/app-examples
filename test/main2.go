package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Annotation struct {
	Key   string
	Value interface{}
}

type Annotations []Annotation

func (a *Annotations) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawAnnotations []interface{}
	if err := unmarshal(&rawAnnotations); err != nil {
		return fmt.Errorf("failed to unmarshal annotations: %w", err)
	}

	var annotations []Annotation

	for _, raw := range rawAnnotations {
		switch item := raw.(type) {
		case string:
			// Handle "key=value" style annotations
			parts := strings.SplitN(item, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid annotation format: %s", item)
			}
			annotations = append(annotations, Annotation{Key: parts[0], Value: parts[1]})

		case map[interface{}]interface{}:
			// Handle annotations as map[interface{}]interface{}
			for k, v := range item {
				keyStr, ok := k.(string)
				if !ok {
					return fmt.Errorf("annotation key is not a string: %v", k)
				}

				jsonValue, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to marshal value for key %s: %w", keyStr, err)
				}
				annotations = append(annotations, Annotation{Key: keyStr, Value: string(jsonValue)})
			}

		case map[string]interface{}:
			// Handle annotations as map[string]interface{}
			for k, v := range item {
				jsonValue, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to marshal value for key %s: %w", k, err)
				}
				annotations = append(annotations, Annotation{Key: k, Value: string(jsonValue)})
			}

		default:
			return fmt.Errorf("unsupported annotation type: %T", item)
		}
	}

	*a = annotations
	return nil
}


// AnnotationsMap processes annotations and returns a map with string keys and interface{} values
func (a Annotations) AnnotationsMap() map[string]interface{} {
	annotations := make(map[string]interface{})

	for _, annotation := range a {
		value := annotation.Value

		// Attempt to parse the value as JSON and convert to YAML
		var jsonObject interface{}
		if json.Unmarshal([]byte(fmt.Sprintf("%v", value)), &jsonObject) == nil {
			// Successfully parsed JSON; convert to YAML
			yamlBytes, err := yaml.Marshal(jsonObject)
			if err != nil {
				log.Printf("Error marshaling JSON to YAML for key %s: %v", annotation.Key, err)
				annotations[annotation.Key] = value // Fallback to original value
			} else {
				annotations[annotation.Key] = string(yamlBytes)
			}
		} else {
			// Value is not JSON, handle as string or simple value
			annotations[annotation.Key] = value
		}
	}

	return annotations
}

// Convert multiline string or YAML into a slice of strings, one per line
func splitIntoLines(value interface{}) []string {
	var lines []string
	var valueStr string

	// Convert the value to string if it's not already
	switch v := value.(type) {
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		// Marshal the interface value to a string (YAML or JSON)
		bytes, err := yaml.Marshal(v)
		if err != nil {
			log.Printf("Error marshaling interface to YAML: %v", err)
			valueStr = fmt.Sprintf("%v", v)
		} else {
			valueStr = string(bytes)
		}
	}

	// Split based on newlines
	for _, line := range strings.Split(valueStr, "\n") {
		lines = append(lines, line)
	}
	return lines
}

func main() {

	yamlData := `
annotations:
  - "key1=value1"
  - key2: "value2"
  - key3: |
      this is it
  - ad.datadoghq.com/test.checks: |
      {   
          "my-test": {
          "init_config": {
              "is_jmx": true,
              "collect_default_metrics": true
          },
          "instances": [{
              "host": "localhost",
              "port": "1234"
          }]
      }
      }
`

	// Parse the YAML into Annotations
	var data struct {
		Annotations Annotations `yaml:"annotations"`
	}

	if err := yaml.Unmarshal([]byte(yamlData), &data); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// fmt.Println(data)

	// annotations := Annotations{
	// 	{Key: "key1", Value: "value1"},
	// 	{Key: "key2", Value: "value2"},
	// 	{Key: "key3", Value: map[string]interface{}{"my-test": map[string]interface{}{
	// 		"init_config": map[string]interface{}{
	// 			"is_jmx":                  true,
	// 			"collect_default_metrics": true,
	// 		},
	// 		"instances": []interface{}{
	// 			map[string]interface{}{"host": "localhost", "port": "1234"},
	// 		},
	// 	}}},
	// }

	annotationsMap := data.Annotations.AnnotationsMap()
	fmt.Println(annotationsMap)

	// Prepare the template
	tmplStr := `Annotations:
{{- range $key, $value := . }}
  - {{ $key }}:
    {{- $lines := splitIntoLines $value }}
    {{- range $lines }}
      {{ . }}
    {{- end }}
{{- end }}
`

	// Register the custom function to split string into lines
	funcMap := template.FuncMap{
		"splitIntoLines": splitIntoLines,
	}

	tmpl, err := template.New("annotations").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Use os.Stdout as the writer for the template output
	err = tmpl.ExecuteTemplate(os.Stdout, "annotations", annotationsMap)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}
