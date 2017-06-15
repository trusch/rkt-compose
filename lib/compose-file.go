package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ComposeFile struct {
	Name     string
	Cpu      string
	Memory   string
	Networks []string
	Manifest PodManifest
}

type PodManifest struct {
	Apps            []*RuntimeApp         `json:"apps"`
	Volumes         []types.Volume        `json:"volumes"`
	Isolators       []types.Isolator      `json:"isolators"`
	Annotations     types.Annotations     `json:"annotations"`
	Ports           []types.ExposedPort   `json:"ports"`
	UserAnnotations types.UserAnnotations `json:"userAnnotations,omitempty"`
	UserLabels      types.UserLabels      `json:"userLabels,omitempty"`
}

type RuntimeApp struct {
	Name           types.ACName      `json:"name"`
	Image          RuntimeImage      `json:"image"`
	App            *App              `json:"app,omitempty"`
	ReadOnlyRootFS bool              `json:"readOnlyRootFS,omitempty"`
	Mounts         []schema.Mount    `json:"mounts,omitempty"`
	Annotations    types.Annotations `json:"annotations,omitempty"`
}

type App struct {
	Exec              types.Exec            `json:"exec"`
	EventHandlers     []types.EventHandler  `json:"eventHandlers,omitempty"`
	User              string                `json:"user"`
	Group             string                `json:"group"`
	SupplementaryGIDs []int                 `json:"supplementaryGIDs,omitempty"`
	WorkingDirectory  string                `json:"workingDirectory,omitempty"`
	Environment       types.Environment     `json:"environment,omitempty"`
	MountPoints       []types.MountPoint    `json:"mountPoints,omitempty"`
	Ports             []types.Port          `json:"ports,omitempty"`
	Isolators         types.Isolators       `json:"isolators,omitempty"`
	UserAnnotations   types.UserAnnotations `json:"userAnnotations,omitempty"`
	UserLabels        types.UserLabels      `json:"userLabels,omitempty"`
}

type RuntimeImage struct {
	Name   string       `json:"name,omitempty"`
	ID     types.Hash   `json:"id"`
	Labels types.Labels `json:"labels,omitempty"`
}

func NewComposeFile(path string) (*ComposeFile, error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	composeFile := &ComposeFile{}
	if err := yaml.Unmarshal(bs, composeFile); err != nil {
		return nil, err
	}
	if len(composeFile.Networks) == 0 {
		composeFile.Networks = []string{"default"}
	}
	return composeFile, nil
}

func (composeFile *ComposeFile) FetchImages() error {
	log.Print("fetch images...")
	for _, app := range composeFile.Manifest.Apps {
		if app.Image.ID.Empty() {
			url := app.Image.Name
			for _, label := range app.Image.Labels {
				url += fmt.Sprintf(",%v=%v", label.Name, label.Value)
			}
			args := []string{"fetch"}
			if strings.HasPrefix(url, "docker://") {
				args = append(args, "--insecure-options=image")
			}
			args = append(args, url)
			log.Printf("fetching image %v...", url)
			cmd := exec.Command("rkt", args...)
			buf := &bytes.Buffer{}
			cmd.Stdout = buf
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			hash, err := types.NewHash(strings.Trim(buf.String(), "\n"))
			if err != nil {
				return err
			}
			app.Image.ID = *hash
			log.Printf("fetched image %v with id %v.", url, app.Image.ID.String())
		}
	}
	return nil
}

func (composeFile *ComposeFile) GetAppcPodManifest() (*schema.PodManifest, error) {
	ver, _ := types.NewSemVer("0.8.10")
	result := &schema.PodManifest{
		ACKind:          types.ACKind("PodManifest"),
		ACVersion:       *ver,
		Volumes:         composeFile.Manifest.Volumes,
		Isolators:       composeFile.Manifest.Isolators,
		Annotations:     composeFile.Manifest.Annotations,
		Ports:           composeFile.Manifest.Ports,
		UserAnnotations: composeFile.Manifest.UserAnnotations,
		UserLabels:      composeFile.Manifest.UserLabels,
		Apps:            make(schema.AppList, len(composeFile.Manifest.Apps)),
	}
	for idx, app := range composeFile.Manifest.Apps {
		name, err := types.NewACIdentifier(app.Image.Name)
		if err != nil {
			name, _ = types.NewACIdentifier("")
		}
		result.Apps[idx] = schema.RuntimeApp{
			Name: app.Name,
			Image: schema.RuntimeImage{
				Name:   name,
				ID:     app.Image.ID,
				Labels: app.Image.Labels,
			},
			App: &types.App{
				Exec:              app.App.Exec,
				EventHandlers:     app.App.EventHandlers,
				User:              app.App.User,
				Group:             app.App.Group,
				SupplementaryGIDs: app.App.SupplementaryGIDs,
				WorkingDirectory:  app.App.WorkingDirectory,
				Environment:       app.App.Environment,
				MountPoints:       app.App.MountPoints,
				Ports:             app.App.Ports,
				Isolators:         app.App.Isolators,
				UserAnnotations:   app.App.UserAnnotations,
				UserLabels:        app.App.UserLabels,
			},
			ReadOnlyRootFS: app.ReadOnlyRootFS,
			Mounts:         app.Mounts,
			Annotations:    app.Annotations,
		}
		if app.App.User == "" {
			result.Apps[idx].App.User = "0"
		}
		if app.App.Group == "" {
			result.Apps[idx].App.Group = "0"
		}
	}
	if composeFile.Cpu != "" {
		cpuIso, err := types.NewResourceCPUIsolator(composeFile.Cpu, composeFile.Cpu)
		if err != nil {
			return nil, err
		}
		result.Isolators = append(result.Isolators, cpuIso.AsIsolator())
	}

	if composeFile.Memory != "" {
		memIso, err := types.NewResourceMemoryIsolator(composeFile.Memory, composeFile.Memory)
		if err != nil {
			return nil, err
		}
		result.Isolators = append(result.Isolators, memIso.AsIsolator())
	}
	return result, nil
}

func (composeFile *ComposeFile) Prepare(output string) error {
	if err := composeFile.FetchImages(); err != nil {
		return err
	}
	log.Print("generate pod-manifest...")
	manifest, err := composeFile.GetAppcPodManifest()
	if err != nil {
		return err
	}
	bs, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	log.Printf("write pod-manifest to %v...", output)
	defer log.Printf("pod-manifest successfully written to %v.", output)
	return ioutil.WriteFile(output, bs, 0644)
}
