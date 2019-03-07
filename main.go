package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/armon/circbuf"
	"github.com/tidwall/gjson"
)

const (
	tfplanExt            = "tfplan"
	tfplanStdoutFilename = "terraform-plan.stdout"
	tfplanJSONFilename   = "terraform-plan.json"
)

// Resource is a helper type
type Resource struct {
	Action string
	Type   string
	Qualifier
}

// Qualifier represents a specific resource meta type
type Qualifier string

// List of available qualifiers.
const (
	QualifierResource     Qualifier = "resource"
	QualifierDataResource Qualifier = "data"
)

// NewResource is a Constructor for Resource
func NewResource(action string, t string, q Qualifier) *Resource {
	return &Resource{Action: action, Type: t, Qualifier: q}
}

func (r *Resource) ToString() string {
	return fmt.Sprintf(`%v`, r)
}

func getStateResources() []*Resource {
	// need to add read only permissions to all exisiting resources
	cmdOutput := execCmd("terraform state list")
	var resources []*Resource
	searchExpression := fmt.Sprintf(`%s_(.*?)[^.]+`, "google")
	re := regexp.MustCompile(searchExpression)
	resourceMatches := re.FindAllString(cmdOutput, -1)
	for _, r := range resourceMatches {
		r = strings.Split(r, ".")[0]
		resources = append(resources, NewResource("read", r, QualifierResource))
	}
	return resources
}

func main() {
	// get flags
	dirPtr := flag.String("dir", ".", "dir")
	permissionsFilenamePtr := flag.String("permissions", "permissions.json", "permissions")
	flag.Parse()
	dir := *dirPtr
	permissionsFilename := *permissionsFilenamePtr

	// change to folder where configuration code is in
	cwd, _ := os.Getwd()
	os.Chdir(dir)

	// execute a terraform plan and save the file in a temporary file
	command := fmt.Sprintf("terraform plan > %s", tfplanStdoutFilename)
	execCmd(command)

	// convert the plan into a json file
	command = fmt.Sprintf("parse-terraform-plan --pretty -i %s -o %s", tfplanStdoutFilename, tfplanJSONFilename)
	execCmd(command)

	// get a list of Resource objects from the state file and the plan
	resources := getStateResources()
	resources = append(resources, parseTerraformPlan(tfplanJSONFilename)...)
	fmt.Printf("\n#############################################\n")
	fmt.Printf("  Terraform Actions to Take: \n")
	fmt.Printf("#############################################\n\n")
	for _, r := range resources {
		fmt.Printf("%s\n", r.ToString())
	}
	//clean up some files and go back to cwd
	os.Remove(tfplanStdoutFilename)
	os.Remove(tfplanJSONFilename)
	os.Chdir(cwd)

	// create a permissions list based on the resources from the plan
	dat, _ := ioutil.ReadFile(permissionsFilename)

	json := string(dat)
	setPermissions := map[string]bool{}
	for _, r := range resources {
		t := reflect.ValueOf(r.Qualifier).String()
		permissions := gjson.Get(json, t+"."+r.Type+"."+r.Action).Array()
		for _, p := range permissions {
			setPermissions[p.String()] = true
		}
	}
	permissions := make([]string, 0, len(setPermissions))
	for k := range setPermissions {
		permissions = append(permissions, k)
	}

	// helpful message to user
	sort.Strings(permissions)
	fmt.Printf("\n#############################################\n")
	fmt.Printf("  Required permissions for deployment role: \n")
	fmt.Printf("#############################################\n\n")
	for _, p := range permissions {
		fmt.Printf("%s\n", p)
	}
}

func parseTerraformPlan(planPath string) []*Resource {
	dat, _ := ioutil.ReadFile(planPath)
	json := string(dat)
	result := gjson.GetMany(json, "changedResources", "changedDataSources")
	// Parse json into a convenient structure
	var resources []*Resource
	for _, r := range result {
		for _, value := range r.Array() {
			action := value.Get("action").String()
			//naming is weird
			if action == "replace" {
				action = "update"
			}
			t := value.Get("type").String()
			path := value.Get("path").String()
			qualifier := QualifierResource
			if strings.HasPrefix(path, "data") {
				qualifier = QualifierDataResource
			}
			resources = append(resources, NewResource(action, t, qualifier))
		}
	}
	return resources
}

func execCmd(command string) string {
	const maxBufSize = 16 * 1024
	// Execute the command using a shell
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}
	cmd := exec.Command(shell, flag, command)
	stdout, _ := circbuf.NewBuffer(maxBufSize)
	stderr, _ := circbuf.NewBuffer(maxBufSize)
	cmd.Stderr = io.Writer(stderr)
	cmd.Stdout = io.Writer(stdout)
	cmd.Run()
	return stdout.String()
}
