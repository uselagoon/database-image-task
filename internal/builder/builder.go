package builder

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/uselagoon/machinery/utils/variables"
)

func fixServiceName(str string) string {
	replaceHyphen := strings.ReplaceAll(str, "-", "_")
	return strings.ToUpper(replaceHyphen)
}

type Builder struct {
	DockerComposeServiceName      string `json:"serviceName"`
	FixedDockerComposeServiceName string `json:"fixedServiceName"`
	SourceImageName               string `json:"sourceImage"`
	CleanImageName                string `json:"cleanImage"`
	ResultImageName               string `json:"resultImageName"`
	ResultImageTag                string `json:"resultImageTag"`
	ResultImageDatabaseName       string `json:"resultImageDatabaseName"`
	RegistryUsername              string `json:"registryUsername"`
	RegistryPassword              string `json:"registryPassword"`
	RegistryHost                  string `json:"registryHost"`
	RegistryOrganization          string `json:"registryOrganization"`
	DockerHost                    string `json:"dockerHost"`
	PushTags                      string `json:"pushTags"`
	MTKYAML                       string `json:"mtkYAML"`
	Debug                         bool   `json:"debug,omitemtpy"`
	MTK                           MTK    `json:"mtk"`
}

type MTK struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func generateBuildValues(vars []variables.LagoonEnvironmentVariable) Builder {
	debugStr := checkVariable("BUILDER_IMAGE_DEBUG", variables.GetEnv("BUILDER_IMAGE_DEBUG", ""), vars)
	debug, _ := strconv.ParseBool(debugStr)
	build := Builder{
		DockerComposeServiceName:      checkVariable("BUILDER_DOCKER_COMPOSE_SERVICE_NAME", variables.GetEnv("BUILDER_DOCKER_COMPOSE_SERVICE_NAME", "mariadb"), vars),
		FixedDockerComposeServiceName: fixServiceName(checkVariable("BUILDER_DOCKER_COMPOSE_SERVICE_NAME", variables.GetEnv("BUILDER_DOCKER_COMPOSE_SERVICE_NAME", "mariadb"), vars)),
		SourceImageName:               checkVariable("BUILDER_IMAGE_NAME", variables.GetEnv("BUILDER_IMAGE_NAME", "mariadb:10.6"), vars),
		CleanImageName:                checkVariable("BUILDER_CLEAN_IMAGE_NAME", variables.GetEnv("BUILDER_CLEAN_IMAGE_NAME", "uselagoon/mariadb-10.6-drupal:latest"), vars),
		ResultImageName:               checkVariable("BUILDER_BACKUP_IMAGE_NAME", variables.GetEnv("BUILDER_BACKUP_IMAGE_NAME", "${project}/${environment}"), vars),
		ResultImageTag:                checkVariable("BUILDER_BACKUP_IMAGE_TAG", variables.GetEnv("BUILDER_BACKUP_IMAGE_TAG", ""), vars),
		ResultImageDatabaseName:       checkVariable("BUILDER_BACKUP_IMAGE_DATABASE_NAME", variables.GetEnv("BUILDER_BACKUP_IMAGE_DATABASE_NAME", ""), vars),
		RegistryUsername:              checkVariable("BUILDER_REGISTRY_USERNAME", variables.GetEnv("BUILDER_REGISTRY_USERNAME", ""), vars),
		RegistryPassword:              checkVariable("BUILDER_REGISTRY_PASSWORD", variables.GetEnv("BUILDER_REGISTRY_PASSWORD", ""), vars),
		RegistryHost:                  checkVariable("BUILDER_REGISTRY_HOST", variables.GetEnv("BUILDER_REGISTRY_HOST", ""), vars),
		RegistryOrganization:          checkVariable("BUILDER_REGISTRY_ORGANIZATION", variables.GetEnv("BUILDER_REGISTRY_ORGANIZATION", ""), vars),
		DockerHost:                    checkVariable("BUILDER_DOCKER_HOST", variables.GetEnv("BUILDER_DOCKER_HOST", "docker-host.lagoon-image-builder.svc"), vars),
		PushTags:                      checkVariable("BUILDER_PUSH_TAGS", variables.GetEnv("BUILDER_PUSH_TAGS", "both"), vars),
		MTKYAML:                       checkVariable("BUILDER_MTK_YAML_BASE64", variables.GetEnv("BUILDER_MTK_YAML_BASE64", ""), vars),
		Debug:                         debug,
	}
	return build
}

// Run will generateValues then output the resulting payload as JSON for the builder script to use
func Run() error {
	vals, err := generateValues()
	if err != nil {
		return err
	}
	b, err := json.Marshal(vals)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// generateValues will get the build values, and then generate the values for MTK
// it also handles scanning for readreplicas if available and parsing the image pattern
func generateValues() (Builder, error) {
	vars := readVariables()
	build := generateBuildValues(vars)
	mtk := MTK{}
	var err error
	mtk.Host, err = calculateMTKVariable("HOSTNAME", build, vars)
	if err != nil {
		return build, err
	}
	mtk.Username, err = calculateMTKVariable("USERNAME", build, vars)
	if err != nil {
		return build, err
	}
	mtk.Password, err = calculateMTKVariable("PASSWORD", build, vars)
	if err != nil {
		return build, err
	}
	mtk.Database, err = calculateMTKVariable("DATABASE", build, vars)
	if err != nil {
		return build, err
	}
	// use a readreplica if one exists
	readReplicas := variables.GetEnv(fmt.Sprintf("%s_READREPLICA_HOSTS", build.FixedDockerComposeServiceName), mtk.Host)
	rr := strings.Split(readReplicas, ",")
	if rr != nil {
		mtk.Host = rr[0]
	}
	build.MTK = mtk
	build.ResultImageName = imagePatternParser(build.ResultImageName, build)
	return build, nil
}

// calculateMTKVariable takes the build vars and environment variables and scans for the necessary variables
func calculateMTKVariable(name string, build Builder, vars []variables.LagoonEnvironmentVariable) (string, error) {
	// support new raw basic `MTK_*` variable
	fVar := fmt.Sprintf("MTK_%s", name)
	sVar := fmt.Sprintf("BUILDER_%s", fVar)
	sVarVal := checkVariable(sVar, "", vars)
	if sVarVal != "" {
		return sVarVal, nil
	}

	// fall back to support pre-existing MTK_DUMP_*
	fVar = fmt.Sprintf("MTK_DUMP_%s", name)
	sVar = fmt.Sprintf("BUILDER_%s", fVar)
	sVarVal = checkVariable(sVar, "", vars)
	if sVarVal != "" {
		return sVarVal, nil
	}

	// support new MTK_*_NAME
	// get the name of the lookup variable
	sVar = checkVariable(fmt.Sprintf("MTK_%s_NAME", name), "", vars)
	if sVar != "" {
		// check that this variable exists with a value
		sVarVal = checkVariable(sVar, "", vars)
		if sVarVal != "" {
			return sVarVal, nil
		}
		return "", fmt.Errorf("no variable found for %s", sVar)
	}

	// fall back to the default servicename variable
	sVarVal = variables.GetEnv(fmt.Sprintf("%s_%s", build.FixedDockerComposeServiceName, name), sVarVal)
	return sVarVal, nil
}

// imagePatternParser parses the image pattern
func imagePatternParser(pattern string, build Builder) string {
	pattern = strings.Replace(pattern, "${database}", build.MTK.Database, 1)
	pattern = strings.Replace(pattern, "${service}", build.DockerComposeServiceName, 1)
	pattern = strings.Replace(pattern, "${registry}", build.RegistryHost, 1)
	pattern = strings.Replace(pattern, "${organization}", build.RegistryOrganization, 1)
	pattern = strings.Replace(pattern, "${project}", variables.GetEnv("LAGOON_PROJECT", ""), 1)
	pattern = strings.Replace(pattern, "${environment}", variables.GetEnv("LAGOON_ENVIRONMENT", ""), 1)
	return pattern
}
