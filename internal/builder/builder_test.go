package builder

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/machinery/utils/variables"
)

func Test_generateValues(t *testing.T) {
	type args struct {
		projectVars []variables.LagoonEnvironmentVariable
		envVars     []variables.LagoonEnvironmentVariable
		setVars     []EnvironmentVariable
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        Builder
	}{
		{
			name:        "test1",
			description: "check mtk values for existing MTK_DUMP prefixed variables",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_HOSTNAME", Value: "dbhost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_USERNAME", Value: "dbuser", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_PASSWORD", Value: "dbpass", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_DATABASE", Value: "dbname", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "lagpro/lagenv",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbhost",
					Username: "dbuser",
					Password: "dbpass",
					Database: "dbname",
				},
			},
		},
		{
			name:        "test2",
			description: "check mtk values for newer MTK_ prefixed variables",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_BACKUP_IMAGE_NAME", Value: "${registry}/${service}-data", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "BUILDER_MTK_HOSTNAME", Value: "dbhost", Scope: "global"},
					{Name: "BUILDER_MTK_USERNAME", Value: "dbuser", Scope: "global"},
					{Name: "BUILDER_MTK_PASSWORD", Value: "dbpass", Scope: "global"},
					{Name: "BUILDER_MTK_DATABASE", Value: "dbname", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "reghost/mariadb-data",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbhost",
					Username: "dbuser",
					Password: "dbpass",
					Database: "dbname",
				},
			},
		},
		{
			name:        "test3",
			description: "check mtk values for existing MTK_DUMP prefixed variables with readreplicas",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_BACKUP_IMAGE_NAME", Value: "${registry}/${service}-data", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_HOSTNAME", Value: "dbhost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_USERNAME", Value: "dbuser", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_PASSWORD", Value: "dbpass", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_DATABASE", Value: "dbname", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "MARIADB_READREPLICA_HOSTS", Value: "dbrrhost1,dbrrhost2"},
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "reghost/mariadb-data",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbrrhost1",
					Username: "dbuser",
					Password: "dbpass",
					Database: "dbname",
				},
			},
		},
		{
			name:        "test4",
			description: "check mtk values for name scoped variables",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_BACKUP_IMAGE_NAME", Value: "${registry}/${service}-data", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "MTK_HOSTNAME_NAME", Value: "DB_HOSTNAME_CENTRAL", Scope: "global"},
					{Name: "MTK_USERNAME_NAME", Value: "DB_USERNAME_CENTRAL", Scope: "global"},
					{Name: "MTK_PASSWORD_NAME", Value: "DB_PASSWORD_CENTRAL", Scope: "global"},
					{Name: "MTK_DATABASE_NAME", Value: "DB_DATABASE_CENTRAL", Scope: "global"},
					{Name: "DB_HOSTNAME_CENTRAL", Value: "dbhostcentral", Scope: "global"},
					{Name: "DB_USERNAME_CENTRAL", Value: "dbusercentral", Scope: "global"},
					{Name: "DB_PASSWORD_CENTRAL", Value: "dbpasscentral", Scope: "global"},
					{Name: "DB_DATABASE_CENTRAL", Value: "dbnamecentral", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "reghost/mariadb-data",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbhostcentral",
					Username: "dbusercentral",
					Password: "dbpasscentral",
					Database: "dbnamecentral",
				},
			},
		},
		{
			name:        "test5",
			description: "check mtk values for default service provided values",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_BACKUP_IMAGE_NAME", Value: "${registry}/${service}-data", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "MARIADB_HOSTNAME", Value: "mariadbhost"},
					{Name: "MARIADB_USERNAME", Value: "mariadbuser"},
					{Name: "MARIADB_PASSWORD", Value: "mariadbpass"},
					{Name: "MARIADB_DATABASE", Value: "mariadbdbname"},
					{Name: "MARIADB_READREPLICA_HOSTS", Value: "dbrrhost1,dbrrhost2"},
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "reghost/mariadb-data",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbrrhost1",
					Username: "mariadbuser",
					Password: "mariadbpass",
					Database: "mariadbdbname",
				},
			},
		},
		{
			name:        "test6",
			description: "check mtk values for existing MTK_DUMP prefixed variables and debug variable",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_HOSTNAME", Value: "dbhost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_USERNAME", Value: "dbuser", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_PASSWORD", Value: "dbpass", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_DATABASE", Value: "dbname", Scope: "global"},
					{Name: "BUILDER_IMAGE_DEBUG", Value: "true", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mariadb:10.6",
				CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
				ResultImageDatabaseName:       "drupal",
				ResultImageName:               "lagpro/lagenv",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				Debug:                         true,
				DatabaseType:                  "mariadb",
				MTK: MTK{
					Host:     "dbhost",
					Username: "dbuser",
					Password: "dbpass",
					Database: "dbname",
				},
			},
		},
		{
			name:        "test7",
			description: "same as test1 except mysql",
			args: args{
				envVars: []variables.LagoonEnvironmentVariable{
					{Name: "BUILDER_BACKUP_IMAGE_TYPE", Value: "mysql", Scope: "global"},
					{Name: "BUILDER_DOCKER_COMPOSE_SERVICE_NAME", Value: "mariadb", Scope: "global"},
					{Name: "BUILDER_REGISTRY_USERNAME", Value: "reguser", Scope: "global"},
					{Name: "BUILDER_REGISTRY_PASSWORD", Value: "regpass", Scope: "global"},
					{Name: "BUILDER_REGISTRY_HOST", Value: "reghost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_HOSTNAME", Value: "dbhost", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_USERNAME", Value: "dbuser", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_PASSWORD", Value: "dbpass", Scope: "global"},
					{Name: "BUILDER_MTK_DUMP_DATABASE", Value: "dbname", Scope: "global"},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: Builder{
				DockerComposeServiceName:      "mariadb",
				FixedDockerComposeServiceName: "MARIADB",
				SourceImageName:               "mysql:8.0.41-oracle",
				CleanImageName:                "uselagoon/mysql-8.0:latest",
				ResultImageDatabaseName:       "lagoon",
				ResultImageName:               "lagpro/lagenv",
				DockerHost:                    "docker-host.lagoon-image-builder.svc",
				PushTags:                      "both",
				RegistryUsername:              "reguser",
				RegistryPassword:              "regpass",
				RegistryHost:                  "reghost",
				DatabaseType:                  "mysql",
				MTK: MTK{
					Host:     "dbhost",
					Username: "dbuser",
					Password: "dbpass",
					Database: "dbname",
				},
			},
		},
	}
	for _, tt := range tests {
		envvars, _ := json.Marshal(tt.args.envVars)
		os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", string(envvars))
		for _, envVar := range tt.args.setVars {
			err := os.Setenv(envVar.Name, envVar.Value)
			if err != nil {
				t.Errorf("%v", err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateValues()
			if err != nil {
				t.Errorf("generateValues() = %v", err)
			}
			oJ, _ := json.MarshalIndent(got, "", "  ")
			wJ, _ := json.MarshalIndent(tt.want, "", "  ")
			if string(oJ) != string(wJ) {
				t.Errorf("generateValues() = \n%v", diff.LineDiff(string(oJ), string(wJ)))
			}
		})
		os.Unsetenv("LAGOON_ENVIRONMENT_VARIABLES")
		for _, envVar := range tt.args.setVars {
			err := os.Unsetenv(envVar.Name)
			if err != nil {
				t.Errorf("%v", err)
			}
		}
	}
}

func Test_imagePatternParser(t *testing.T) {
	type args struct {
		pattern string
		build   Builder
		setVars []EnvironmentVariable
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        string
	}{
		{
			name: "test1",
			args: args{
				pattern: "${registry}/${organization}/${project}/${service}-data",
				build: Builder{
					DockerComposeServiceName:      "mariadb",
					FixedDockerComposeServiceName: "MARIADB",
					SourceImageName:               "mariadb:10.6",
					CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
					ResultImageName:               "backup/image",
					DockerHost:                    "docker-host.lagoon-image-builder.svc",
					PushTags:                      "both",
					RegistryUsername:              "reguser",
					RegistryPassword:              "regpass",
					RegistryHost:                  "reghost",
					RegistryOrganization:          "regorg",
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: "reghost/regorg/lagpro/mariadb-data",
		},
		{
			name: "test2",
			args: args{
				pattern: "${registry}/${project}/${environment}/${service}-data",
				build: Builder{
					DockerComposeServiceName:      "mariadb",
					FixedDockerComposeServiceName: "MARIADB",
					SourceImageName:               "mariadb:10.6",
					CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
					ResultImageName:               "backup/image",
					DockerHost:                    "docker-host.lagoon-image-builder.svc",
					PushTags:                      "both",
					RegistryUsername:              "reguser",
					RegistryPassword:              "regpass",
					RegistryHost:                  "reghost",
					RegistryOrganization:          "regorg",
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: "reghost/lagpro/lagenv/mariadb-data",
		},
		{
			name: "test3",
			args: args{
				pattern: "${registry}/${project}/${service}-data",
				build: Builder{
					DockerComposeServiceName:      "mariadb",
					FixedDockerComposeServiceName: "MARIADB",
					SourceImageName:               "mariadb:10.6",
					CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
					ResultImageName:               "backup/image",
					DockerHost:                    "docker-host.lagoon-image-builder.svc",
					PushTags:                      "both",
					RegistryUsername:              "reguser",
					RegistryPassword:              "regpass",
					RegistryHost:                  "reghost",
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: "reghost/lagpro/mariadb-data",
		},
		{
			name:        "test4",
			description: "Check whether $database works",
			args: args{
				pattern: "${organization}/database-mysql-${project}-${environment}-${database}",
				build: Builder{
					DockerComposeServiceName:      "mariadb",
					FixedDockerComposeServiceName: "MARIADB",
					SourceImageName:               "mariadb:10.6",
					CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
					ResultImageName:               "backup/image",
					DockerHost:                    "docker-host.lagoon-image-builder.svc",
					PushTags:                      "both",
					RegistryUsername:              "reguser",
					RegistryPassword:              "regpass",
					RegistryHost:                  "reghost",
					RegistryOrganization:          "regorg",
					MTK: MTK{
						Database: "test_database_name",
					},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: "regorg/database-mysql-lagpro-lagenv-test_database_name",
		},
		{
			name:        "test5",
			description: "Check whether removal of double special characters works",
			args: args{
				pattern: "${organization}/database-mysql-${project}-${environment}-${database}",
				build: Builder{
					DockerComposeServiceName:      "mariadb",
					FixedDockerComposeServiceName: "MARIADB",
					SourceImageName:               "mariadb:10.6",
					CleanImageName:                "uselagoon/mariadb-10.6-drupal:latest",
					ResultImageName:               "backup/image",
					DockerHost:                    "docker-host.lagoon-image-builder.svc",
					PushTags:                      "both",
					RegistryUsername:              "reguser",
					RegistryPassword:              "regpass",
					RegistryHost:                  "reghost",
					RegistryOrganization:          "regorg",
					MTK: MTK{
						Database: "test_database__name!!",
					},
				},
				setVars: []EnvironmentVariable{
					{Name: "LAGOON_PROJECT", Value: "lagpro"},
					{Name: "LAGOON_ENVIRONMENT", Value: "lagenv"},
				},
			},
			want: "regorg/database-mysql-lagpro-lagenv-test_database_name",
		},
	}
	for _, tt := range tests {
		for _, envVar := range tt.args.setVars {
			err := os.Setenv(envVar.Name, envVar.Value)
			if err != nil {
				t.Errorf("%v", err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := imagePatternParser(tt.args.pattern, tt.args.build); got != tt.want {
				t.Errorf("imagePatternParser() = %v, want %v", got, tt.want)
			}
		})
		for _, envVar := range tt.args.setVars {
			err := os.Unsetenv(envVar.Name)
			if err != nil {
				t.Errorf("%v", err)
			}
		}
	}
}
