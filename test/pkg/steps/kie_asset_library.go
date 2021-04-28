// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package steps

import (
	"os"

	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/steps/mappers"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// TODO
// Should be externalized to config.go ?
const (
	KieAssetLibraryGitRepositoryURI    = "https://github.com/jstastny-cz/kie-asset-library-poc"
	KieAssetLibraryGitRepositoryBranch = "main"
)

// registerMavenSteps register all existing Maven steps
func registerKieAssetLibrarySteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step("^Project kie-asset-library is cloned$", data.projectKieAssetLibraryIsCloned)
	ctx.Step("^Project kie-asset-library is built by maven with configuration:$", data.projectKieAssetLibraryIsBuiltByMavenWithConfiguration)
	ctx.Step("^Project \"([^\"]*)\" is generated in temporary folder$", data.projectIsGeneratedInTemporaryFolder)
	ctx.Step("^Project \"([^\"]*)\" is built from temporary folder by maven$", data.projectIsBuiltFromTemporaryFolderByMaven)
	ctx.Step("^Project \"([^\"]*)\" assets are re-marshalled by VS Code$", data.projectAssetsAreRemarshalledByVsCode)
	ctx.Step(`^Build binary (quarkus|springboot) service "([^"]*)" from kie-asset-library target folder$`, data.deployKieAssetTargetOnOpenshift)
}

func (data *Data) projectKieAssetLibraryIsBuiltByMavenWithConfiguration(table *godog.Table) error {

	mavenConfig := &mappers.MavenCommandConfig{}
	if table != nil && len(table.Rows) > 0 {
		err := mappers.MapMavenCommandConfigTable(table, mavenConfig)
		if err != nil {
			return err
		}
	}

	return data.localPathBuiltByMavenWithProfileAndOptions(data.KieAssetLibraryLocation, mavenConfig)
}

func (data *Data) projectKieAssetLibraryIsCloned() error {
	framework.GetLogger(data.Namespace).Info("Cloning kie-asset-library project", "URI", KieAssetLibraryGitRepositoryURI, "branch", KieAssetLibraryGitRepositoryBranch, "clonedLocation", data.KieAssetLibraryLocation)

	cloneOptions := &git.CloneOptions{
		URL:          KieAssetLibraryGitRepositoryURI,
		SingleBranch: true,
	}

	var err error
	reference := KieAssetLibraryGitRepositoryBranch
	if len(reference) == 0 {
		err = cloneRepository(data.KieAssetLibraryLocation, cloneOptions)
	} else {
		// Try cloning as branch reference
		cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(reference)
		err = cloneRepository(data.KieAssetLibraryLocation, cloneOptions)
		// If branch clone was successful then return, otherwise try other cloning options
		if err == nil {
			return nil
		}

		// If branch cloning failed then try cloning as tag
		cloneOptions.ReferenceName = plumbing.NewTagReferenceName(reference)
		err = cloneRepository(data.KogitoExamplesLocation, cloneOptions)
	}
	return err
}

func (data *Data) projectIsGeneratedInTemporaryFolder(project string) error {
	if _, err := os.Stat(data.KieAssetLibraryLocation + "/kie-assets-library-generate/target/" + project); !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (data *Data) projectIsBuiltFromTemporaryFolderByMaven(project string) error {
	_, errCode := framework.CreateMavenCommand(data.KieAssetLibraryLocation+"/kie-assets-library-generate/target/"+project).
		SkipTests().
		Execute("clean", "install")
	if errCode != nil {
		framework.GetLogger(data.Namespace).Warn(project + " 'mvn clean install' failed due to: " + errCode.Error())
	}
	return errCode
}

func (data *Data) projectAssetsAreRemarshalledByVsCode(project string) error {
	// TO DO
	// output, errCode := framework.CreateCommand("yarn",
	// 	"run",
	// 	"test:it",
	// 	"KIE_VSIX=/home/jomarko/Downloads/KOGITO-4179-plugin-v2.vsix",
	// 	"KIE_PROJECT="+data.KogitoExamplesLocation+"kie-assets-library-generate/target/"+project).
	// 	WithRetry(framework.NumberOfRetries(1)).
	// 	InDirectory("/home/jomarko/redhat/github/jomarko/kie-assets-re-marshaller").Execute()
	// framework.GetLogger(data.Namespace).Info(output)
	// return errCode
	return nil
}

func (data *Data) deployKieAssetTargetOnOpenshift(runtimeType, project string, table *godog.Table) error {
	binaryFolder := data.KieAssetLibraryLocation + "/kie-assets-library-generate/target/" + project

	return data.deployTargetFolderOnOpenshift(runtimeType, project, binaryFolder, table)
}
