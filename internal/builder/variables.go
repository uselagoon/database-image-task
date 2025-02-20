package builder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/uselagoon/machinery/utils/variables"
)

type EnvironmentVariable struct {
	Name  string
	Value string
}

// readVariables reads the lagoon environment and project variables and merges them, environment variables take precedence
func readVariables() []variables.LagoonEnvironmentVariable {
	projectVariables := variables.GetEnv("LAGOON_PROJECT_VARIABLES", "")
	environmentVariables := variables.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", "")

	projectVars := []variables.LagoonEnvironmentVariable{}
	envVars := []variables.LagoonEnvironmentVariable{}
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	return mergeVariables(projectVars, envVars)
}

// mergeVariables will be irrelevant once https://github.com/uselagoon/lagoon/pull/3856 is merged and relased, as it will consolidate variables into one payload
// in the future
func mergeVariables(project, environment []variables.LagoonEnvironmentVariable) []variables.LagoonEnvironmentVariable {
	allVars := []variables.LagoonEnvironmentVariable{}
	existsInEnvironment := false
	// replace any variables from the project with ones from the environment
	// this only modifies ones that exist in project
	for _, pVar := range project {
		add := variables.LagoonEnvironmentVariable{}
		for _, eVar := range environment {
			// internal_system scoped variables are only added to the projects variabled during a build
			// this make sure that any that may exist in the environment variables are not merged
			// and also makes sure that internal_system variables are not replaced by other scopes
			if eVar.Name == pVar.Name && pVar.Scope != "internal_system" && eVar.Scope != "internal_system" {
				existsInEnvironment = true
				add = eVar
			}
		}
		if existsInEnvironment {
			allVars = append(allVars, add)
			existsInEnvironment = false
		} else {
			allVars = append(allVars, pVar)
		}
	}
	// add any that exist in the environment only to the final variables list
	existsInProject := false
	for _, eVar := range environment {
		add := eVar
		for _, aVar := range allVars {
			if eVar.Name == aVar.Name {
				existsInProject = true
			}
		}
		if existsInProject {
			existsInProject = false
		} else {
			allVars = append(allVars, add)
		}
	}
	return allVars
}

// checks the provided environment variables looking for feature flag based variables
func checkFeatureFlag(key string, envVariables []variables.LagoonEnvironmentVariable) string {
	// check for force value
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_FORCE_%s", key)); ok {
		return value
	}
	// check lagoon environment variables
	for _, lVar := range envVariables {
		if strings.Contains(lVar.Name, fmt.Sprintf("LAGOON_FEATURE_FLAG_%s", key)) {
			return lVar.Value
		}
	}
	// return default
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_DEFAULT_%s", key)); ok {
		return value
	}
	// otherwise nothing
	return ""
}

// checkVariable will check the variables from the featureflags, json payload and finally the environment variables
// if none found, falls back to the value provided as the default
func checkVariable(name, defValue string, vars []variables.LagoonEnvironmentVariable) string {
	// check any featureflag variables first
	fflag := checkFeatureFlag(name, vars)
	if fflag != "" {
		return fflag
	}
	// get the JSON_PAYLOAD variable and search it for variables
	jsonPayload := variables.GetEnv("JSON_PAYLOAD", "")
	if jsonPayload != "" {
		jsonBytes, _ := base64.StdEncoding.DecodeString(jsonPayload)
		var payloadData map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &payloadData); err != nil {
			os.Stderr.WriteString(fmt.Sprintf("failed to unsmarshal the supplied JSON payload data, error was: %v\n", err))
			return ""
		}
		if v, ok := payloadData[name]; ok {
			return v.(string)
		}
	}
	// search for the variable in the lagoon env vars
	for _, v := range vars {
		if v.Name == name {
			return v.Value
		}
	}
	// fall back to default provided value
	return defValue
}
