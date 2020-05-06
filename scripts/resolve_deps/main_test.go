package main

import (
	"os"
	"reflect"
	"testing"
)

const (
	pluginModuleFileName     = "testdata/go.mod-plugin"
	mergedModuleFileName     = "testdata/go.mod-merged"
	suggestionModuleFileName = "testdata/go.mod-suggestion"
)

func Test_checkFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "existing_file", args: args{filename: pluginModuleFileName}, wantErr: false},
		{name: "non_existing_file", args: args{filename: "foo.mod"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFile(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("checkFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_readModuleFile(t *testing.T) {
	wantPluginMod := &GoModuleDescriptor{
		Module:  "github.com/solo-io/ext-auth-plugin-implementation",
		Version: "1.14",
		Require: map[string]string{
			"github.com/envoyproxy/go-control-plane": "github.com/envoyproxy/go-control-plane v0.9.1",
			"github.com/solo-io/ext-auth-plugins":    "github.com/solo-io/ext-auth-plugins v0.1.1",
			"github.com/solo-io/go-utils":            "github.com/solo-io/go-utils v0.11.5",
			"go.uber.org/zap":                        "go.uber.org/zap v1.13.0",
			"zz.com/zz_mycompany/library":            "zz.com/zz_mycompany/library v1.0.0",
		},
		Replace: map[string]string{
			"github.com/docker/docker": "github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309",
			"k8s.io/cri-api":           "k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac",
		},
	}
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *GoModuleDescriptor
		wantErr bool
	}{
		{name: "non_existing_file.mod", args: args{"foo.mod"}, wantErr: true},
		{name: "unknown.mod", args: args{"testdata/go.mod-unknown"}, wantErr: true},
		{name: "plugin.mod", args: args{pluginModuleFileName}, want: wantPluginMod, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readModuleFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readModuleFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readModuleFile() \ngot:  %+v \nwant: %+v", got, tt.want)
			}
		})
	}
}

func Test_mergeModules(t *testing.T) {
	type args struct {
		suggestionsModule *GoModuleDescriptor
		pluginModule      *GoModuleDescriptor
	}
	tests := []struct {
		name string
		args args
		want *GoModuleDescriptor
	}{
		{
			name: "plugin-module-retained",
			args: args{
				suggestionsModule: &GoModuleDescriptor{Module: "example"},
				pluginModule:      &GoModuleDescriptor{Module: "foo"},
			},
			want: &GoModuleDescriptor{Module: "foo"},
		},
		{
			name: "plugin-version-overwritten-by-suggestion",
			args: args{
				suggestionsModule: &GoModuleDescriptor{Version: "1.13"},
				pluginModule:      &GoModuleDescriptor{Version: "1.14"},
			},
			want: &GoModuleDescriptor{Version: "1.13"},
		},
		{
			name: "plugin-version-retained",
			args: args{
				suggestionsModule: &GoModuleDescriptor{},
				pluginModule:      &GoModuleDescriptor{Version: "1.14"},
			},
			want: &GoModuleDescriptor{Version: "1.14"},
		},
		{
			name: "plugin-require-replace-added-from-suggestion",
			args: args{
				suggestionsModule: &GoModuleDescriptor{
					Require: map[string]string{"bar": "foo"},
					Replace: map[string]string{"foo": "barfoo"},
				},
				pluginModule: &GoModuleDescriptor{
					Require: map[string]string{},
					Replace: map[string]string{},
				},
			},
			want: &GoModuleDescriptor{
				Require: map[string]string{"bar": "foo"},
				Replace: map[string]string{"foo": "barfoo"},
			},
		},
		{
			name: "plugin-require-replace-overwritten-by-suggestion",
			args: args{
				suggestionsModule: &GoModuleDescriptor{
					Require: map[string]string{"bar": "foobar"},
					Replace: map[string]string{"foo": "barfoo"},
				},
				pluginModule: &GoModuleDescriptor{
					Require: map[string]string{"bar": "foo"},
					Replace: map[string]string{"foo": "barfoobar"},
				},
			},
			want: &GoModuleDescriptor{
				Require: map[string]string{"bar": "foobar"},
				Replace: map[string]string{"foo": "barfoo"},
			},
		},
		{
			name: "plugin-require-replace-same-with-nil-suggestion",
			args: args{
				suggestionsModule: &GoModuleDescriptor{},
				pluginModule: &GoModuleDescriptor{
					Require: map[string]string{},
					Replace: map[string]string{},
				},
			},
			want: &GoModuleDescriptor{
				Require: map[string]string{},
				Replace: map[string]string{},
			},
		},
		{
			name: "plugin-with-nil-replace-overwritten-by-suggestion",
			args: args{
				suggestionsModule: &GoModuleDescriptor{
					Replace: map[string]string{"foo": "barfoo"},
				},
				pluginModule: &GoModuleDescriptor{
					Require: map[string]string{},
				},
			},
			want: &GoModuleDescriptor{
				Require: map[string]string{},
				Replace: map[string]string{"foo": "barfoo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeModules(tt.args.suggestionsModule, tt.args.pluginModule)
			if !reflect.DeepEqual(tt.args.pluginModule, tt.want) {
				t.Errorf("mergeModules() \ngot:  %+v \nwant: %+v", tt.args.pluginModule, tt.want)
			}
		})
	}
}

func Test_read_merge_and_createModuleFile(t *testing.T) {
	pluginModule, err := readModuleFile(pluginModuleFileName)
	if err != nil {
		t.Fatalf("failed during reading %s", pluginModuleFileName)
	}
	suggestionModule, err := readModuleFile(suggestionModuleFileName)
	if err != nil {
		t.Fatalf("failed during reading %s", suggestionModuleFileName)
	}
	t.Run("Merged module", func(t *testing.T) {
		mergeModules(suggestionModule, pluginModule)
		if err := createModuleFile(pluginModule); err != nil {
			t.Errorf("createModuleFile() error = %v", err)
		}
		mergedModule, err := readModuleFile(mergeSuggestedModuleFileName)
		if err != nil {
			t.Errorf("failed during read of %s, %v", mergeSuggestedModuleFileName, err)
		}
		wantMergedModule, err := readModuleFile(mergedModuleFileName)
		if err != nil {
			t.Errorf("failed during read of %s, %v", mergedModuleFileName, err)
		}
		if !reflect.DeepEqual(mergedModule, wantMergedModule) {
			t.Errorf("readModuleFile() \ngot:  %+v \nwant: %+v", mergedModule, wantMergedModule)
		}

		t.Cleanup(func() {
			if !t.Failed() {
				_ = os.Remove(mergeSuggestedModuleFileName)
			}
		})
	})

}
