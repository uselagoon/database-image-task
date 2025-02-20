package builder

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/machinery/utils/variables"
)

func Test_mergeVariables(t *testing.T) {
	type args struct {
		project     []variables.LagoonEnvironmentVariable
		environment []variables.LagoonEnvironmentVariable
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        []variables.LagoonEnvironmentVariable
	}{
		{
			name:        "test1",
			description: "test that a variable defined in the project is overridden by a variable defined at the environment level",
			args: args{
				project: []variables.LagoonEnvironmentVariable{
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "projectdatabase",
						Scope: "global",
					},
				},
				environment: []variables.LagoonEnvironmentVariable{
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "environmentdatabase",
						Scope: "global",
					},
				},
			},
			want: []variables.LagoonEnvironmentVariable{
				{
					Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
					Value: "environmentdatabase",
					Scope: "global",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeVariables(tt.args.project, tt.args.environment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkVariable(t *testing.T) {
	type args struct {
		reqname  string
		defvalue string
		vars     []variables.LagoonEnvironmentVariable
		setVars  []EnvironmentVariable
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        string
	}{
		{
			name:        "test1",
			description: "check that the requested variable is not in the json payload or the environment variables, fall back to the default",
			args: args{
				setVars: []EnvironmentVariable{
					{
						Name:  "JSON_PAYLOAD",
						Value: "",
					},
				},
				vars: []variables.LagoonEnvironmentVariable{
					{
						Name:  "RANDOM_VARIABLE",
						Value: "randomvalue",
						Scope: "global",
					},
				},
				reqname:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
				defvalue: "defaultvalue",
			},
			want: "defaultvalue",
		},
		{
			name:        "test2",
			description: "check that the requested variable is defined as a feature flag from the controller",
			args: args{
				setVars: []EnvironmentVariable{
					{
						Name:  "JSON_PAYLOAD",
						Value: "",
					},
					{
						Name:  "LAGOON_FEATURE_FLAG_DEFAULT_BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "featuredatabase",
					},
				},
				vars: []variables.LagoonEnvironmentVariable{
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "environmentdatabase",
						Scope: "global",
					},
				},
				reqname:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
				defvalue: "defaultvalue",
			},
			want: "featuredatabase",
		},
		{
			name:        "test3",
			description: "check that the requested variable is defined as a feature flag in the environment",
			args: args{
				setVars: []EnvironmentVariable{
					{
						Name:  "JSON_PAYLOAD",
						Value: "",
					},
				},
				vars: []variables.LagoonEnvironmentVariable{
					{
						Name:  "LAGOON_FEATURE_FLAG_BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "featuredatabase",
					},
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "environmentdatabase",
						Scope: "global",
					},
				},
				reqname:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
				defvalue: "defaultvalue",
			},
			want: "featuredatabase",
		},
		{
			name:        "test4",
			description: "check that the variable requested is in the environmentvariables",
			args: args{
				setVars: []EnvironmentVariable{
					{
						Name:  "JSON_PAYLOAD",
						Value: "",
					},
				},
				vars: []variables.LagoonEnvironmentVariable{
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "environmentdatabase",
						Scope: "global",
					},
				},
				reqname:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
				defvalue: "defaultvalue",
			},
			want: "environmentdatabase",
		},
		{
			name:        "test5",
			description: "check that the variable requested is in the json payload",
			args: args{
				setVars: []EnvironmentVariable{
					{
						Name:  "JSON_PAYLOAD",
						Value: genBase64JSONPayload(map[string]string{"BUILDER_DOCKER_COMPOSE_SERVICE_NAME": "jsondatabase"}),
					},
				},
				vars: []variables.LagoonEnvironmentVariable{
					{
						Name:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
						Value: "environmentdatabase",
						Scope: "global",
					},
				},
				reqname:  "BUILDER_DOCKER_COMPOSE_SERVICE_NAME",
				defvalue: "defaultvalue",
			},
			want: "jsondatabase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, envVar := range tt.args.setVars {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			if got := checkVariable(tt.args.reqname, tt.args.defvalue, tt.args.vars); got != tt.want {
				t.Errorf("checkVariable() = %v, want %v", got, tt.want)
			}
			for _, envVar := range tt.args.setVars {
				err := os.Unsetenv(envVar.Name)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
		})
	}
}

func genBase64JSONPayload(p map[string]string) string {
	b, _ := json.Marshal(p)
	return base64.StdEncoding.EncodeToString(b)
}
