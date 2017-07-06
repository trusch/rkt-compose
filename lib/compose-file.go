package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ComposeFile represents a single compose file
type ComposeFile struct {
	Name     string      `json:"name" yaml:"name,omitempty"`
	CPU      string      `json:"cpu" yaml:"cpu,omitempty"`
	Memory   string      `json:"memory" yaml:"memory,omitempty"`
	Networks []string    `json:"networks" yaml:"networks,omitempty"`
	Extra    []string    `json:"extra" yaml:"extra,omitempty"`
	Manifest PodManifest `json:"manifest" yaml:"manifest,omitempty"`
}

// A PodManifest mimics the appc PodManifest but without validation
type PodManifest struct {
	Apps            []*RuntimeApp         `json:"apps" yaml:"apps,omitempty"`
	Volumes         []*Volume             `json:"volumes" yaml:"volumes,omitempty"`
	Isolators       []types.Isolator      `json:"isolators" yaml:"isolators,omitempty"`
	Annotations     types.Annotations     `json:"annotations" yaml:"annotations,omitempty"`
	Ports           []types.ExposedPort   `json:"ports" yaml:"ports,omitempty"`
	UserAnnotations types.UserAnnotations `json:"userAnnotations,omitempty" yaml:"userAnnotations,omitempty"`
	UserLabels      types.UserLabels      `json:"userLabels,omitempty" yaml:"userLabels,omitempty"`
}

// A RuntimeApp mimics the appc RuntimeApp but without validation
type RuntimeApp struct {
	Name           types.ACName      `json:"name" yaml:"name,omitempty"`
	Image          RuntimeImage      `json:"image" yaml:"image,omitempty"`
	App            *App              `json:"app,omitempty" yaml:"app,omitempty"`
	ReadOnlyRootFS bool              `json:"readOnlyRootFS,omitempty" yaml:"readOnlyRootFS,omitempty"`
	Mounts         []schema.Mount    `json:"mounts,omitempty" yaml:"mounts,omitempty"`
	Annotations    types.Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// A App mimics the appc App but without validation
type App struct {
	Exec              types.Exec            `json:"exec" yaml:"exec,omitempty"`
	EventHandlers     []types.EventHandler  `json:"eventHandlers,omitempty" yaml:"eventHandlers,omitempty"`
	User              string                `json:"user" yaml:"user,omitempty"`
	Group             string                `json:"group" yaml:"group,omitempty"`
	SupplementaryGIDs []int                 `json:"supplementaryGIDs,omitempty" yaml:"supplementaryGIDs,omitempty"`
	WorkingDirectory  string                `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	Environment       types.Environment     `json:"environment,omitempty" yaml:"environment,omitempty"`
	MountPoints       []types.MountPoint    `json:"mountPoints,omitempty" yaml:"mountPoints,omitempty"`
	Ports             []types.Port          `json:"ports,omitempty" yaml:"ports,omitempty"`
	Isolators         types.Isolators       `json:"isolators,omitempty" yaml:"isolators,omitempty"`
	UserAnnotations   types.UserAnnotations `json:"userAnnotations,omitempty" yaml:"userAnnotations,omitempty"`
	UserLabels        types.UserLabels      `json:"userLabels,omitempty" yaml:"userLabels,omitempty"`
}

// A RuntimeImage mimics the appc RuntimeImage but without validation
type RuntimeImage struct {
	Name   string       `json:"name,omitempty" yaml:"name,omitempty"`
	ID     types.Hash   `json:"id" yaml:"id,omitempty"`
	Labels types.Labels `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// A Volume mimics the appc Volume but without validation
type Volume struct {
	Name      types.ACName `json:"name" yaml:"name,omitempty"`
	Kind      string       `json:"kind" yaml:"kind,omitempty"`
	Source    string       `json:"source,omitempty" yaml:"source,omitempty"`
	ReadOnly  *bool        `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Recursive *bool        `json:"recursive,omitempty" yaml:"recursive,omitempty"`
	Mode      *string      `json:"mode,omitempty" yaml:"mode,omitempty"`
	UID       *int         `json:"uid,omitempty" yaml:"uid,omitempty"`
	GID       *int         `json:"gid,omitempty" yaml:"gid,omitempty"`
}

// NewComposeFile parses a composefile from disk
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

func (composeFile *ComposeFile) fetchImages() error {
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

// GetAppcPodManifest returns a appc conform manifest for this pod
func (composeFile *ComposeFile) GetAppcPodManifest() (*schema.PodManifest, error) {
	ver, _ := types.NewSemVer("0.8.10")
	volumes := make([]types.Volume, len(composeFile.Manifest.Volumes))
	for idx, vol := range composeFile.Manifest.Volumes {
		volumes[idx] = types.Volume{
			Name:      vol.Name,
			Kind:      vol.Kind,
			Source:    vol.Source,
			ReadOnly:  vol.ReadOnly,
			Recursive: vol.Recursive,
			Mode:      vol.Mode,
			UID:       vol.UID,
			GID:       vol.GID,
		}
	}
	result := &schema.PodManifest{
		ACKind:          types.ACKind("PodManifest"),
		ACVersion:       *ver,
		Volumes:         volumes,
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
	if composeFile.CPU != "" {
		cpuIso, err := types.NewResourceCPUIsolator(composeFile.CPU, composeFile.CPU)
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

func (composeFile *ComposeFile) assertVolumes() error {
	for _, volume := range composeFile.Manifest.Volumes {
		if volume.Kind == "" {
			volume.Kind = "host"
		}
		if volume.Kind == "host" {
			if strings.HasPrefix(volume.Source, "./") {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				volume.Source = filepath.Join(cwd, volume.Source)
			}
			if _, err := os.Stat(volume.Source); err != nil {
				err = os.MkdirAll(volume.Source, 0777)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Prepare fetches images and creates host volume pathes if needed
func (composeFile *ComposeFile) Prepare(output io.Writer) error {
	if err := composeFile.fetchImages(); err != nil {
		return err
	}
	if err := composeFile.assertVolumes(); err != nil {
		return err
	}
	log.Print("generate pod-manifest...")
	manifest, err := composeFile.GetAppcPodManifest()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	err = encoder.Encode(manifest)
	if err != nil {
		return err
	}
	log.Print("pod-manifest successfully written")
	return nil
}
