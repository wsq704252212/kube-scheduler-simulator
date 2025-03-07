package scheduler

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1beta2config "k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	schedConfig "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
)

func Test_convertConfigurationForSimulator(t *testing.T) {
	t.Parallel()

	var nondefaultParallelism int32 = 3
	defaultschedulername := v1.DefaultSchedulerName
	nondefaultschedulername := v1.DefaultSchedulerName + "2"

	var minCandidateNodesPercentage int32 = 20
	var minCandidateNodesAbsolute int32 = 100
	var hardPodAffinityWeight int32 = 2

	type args struct {
		versioned *v1beta2config.KubeSchedulerConfiguration
		port      int
	}
	tests := []struct {
		name    string
		args    args
		want    *config.KubeSchedulerConfiguration
		wantErr bool
	}{
		{
			name: "success with empty-configuration",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{},
				port:      80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with no-disabled plugin",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with empty Profiles",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{},
				port:      80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles and Extenders does not affects result",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "changes of field other than Profiles.Plugins and Extenders does not affects result",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins:       &v1beta2config.Plugins{},
							PluginConfig:  nil,
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
						},
						{
							SchedulerName: &nondefaultschedulername,
							Plugins: &v1beta2config.Plugins{
								Score: v1beta2config.PluginSet{
									Disabled: []v1beta2config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				profile2.Plugins.Score.Enabled = []config.Plugin{
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: 1},
					{Name: "InterPodAffinityWrapped", Weight: 1},
					{Name: "NodeAffinityWrapped", Weight: 1},
					{Name: "PodTopologySpreadWrapped", Weight: 2},
					{Name: "TaintTolerationWrapped", Weight: 1},
				}
				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with multiple profiles and custom-pluginconfig",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							PluginConfig: []v1beta2config.PluginConfig{
								{
									Name: "DefaultPreemption",
									Args: runtime.RawExtension{
										Object: &v1beta2config.DefaultPreemptionArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "DefaultPreemptionArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1beta2",
											},
											MinCandidateNodesPercentage: &minCandidateNodesPercentage,
											MinCandidateNodesAbsolute:   &minCandidateNodesAbsolute,
										},
									},
								},
							},
						},
						{
							SchedulerName: &nondefaultschedulername,
							PluginConfig: []v1beta2config.PluginConfig{
								{
									Name: "InterPodAffinity",
									Args: runtime.RawExtension{
										Object: &v1beta2config.InterPodAffinityArgs{
											TypeMeta: metav1.TypeMeta{
												Kind:       "InterPodAffinityArgs",
												APIVersion: "kubescheduler.config.k8s.io/v1beta2",
											},
											HardPodAffinityWeight: &hardPodAffinityWeight,
										},
									},
								},
							},
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				profile2 := cfg.Profiles[0].DeepCopy()
				profile2.SchedulerName = nondefaultschedulername
				for i := range cfg.Profiles[0].PluginConfig {
					if cfg.Profiles[0].PluginConfig[i].Name == "DefaultPreemption" {
						cfg.Profiles[0].PluginConfig[i] = config.PluginConfig{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{
								MinCandidateNodesPercentage: minCandidateNodesPercentage,
								MinCandidateNodesAbsolute:   minCandidateNodesAbsolute,
							},
						}
					}
					if cfg.Profiles[0].PluginConfig[i].Name == "DefaultPreemptionWrapped" {
						cfg.Profiles[0].PluginConfig[i] = config.PluginConfig{
							Name: "DefaultPreemptionWrapped",
							Args: &config.DefaultPreemptionArgs{
								MinCandidateNodesPercentage: minCandidateNodesPercentage,
								MinCandidateNodesAbsolute:   minCandidateNodesAbsolute,
							},
						}
					}
				}

				for i := range profile2.PluginConfig {
					if profile2.PluginConfig[i].Name == "InterPodAffinity" {
						profile2.PluginConfig[i] = config.PluginConfig{
							Name: "InterPodAffinity",
							Args: &config.InterPodAffinityArgs{
								HardPodAffinityWeight: hardPodAffinityWeight,
							},
						}
					}
					if profile2.PluginConfig[i].Name == "InterPodAffinityWrapped" {
						profile2.PluginConfig[i] = config.PluginConfig{
							Name: "InterPodAffinityWrapped",
							Args: &config.InterPodAffinityArgs{
								HardPodAffinityWeight: hardPodAffinityWeight,
							},
						}
					}
				}

				cfg.Profiles = append(cfg.Profiles, *profile2)
				return &cfg
			}(),
		},
		{
			name: "success with some plugin disabled",
			args: args{
				versioned: &v1beta2config.KubeSchedulerConfiguration{
					Parallelism: &nondefaultParallelism,
					Profiles: []v1beta2config.KubeSchedulerProfile{
						{
							SchedulerName: &defaultschedulername,
							Plugins: &v1beta2config.Plugins{
								Score: v1beta2config.PluginSet{
									Disabled: []v1beta2config.Plugin{
										{
											Name: "ImageLocality",
										},
										{
											Name: "NodeResourcesFit",
										},
									},
								},
							},
						},
					},
					Extenders: []v1beta2config.Extender{
						{
							URLPrefix:      "http://example.com/extender/",
							PreemptVerb:    "PreemptVerb/",
							FilterVerb:     "FilterVerb/",
							PrioritizeVerb: "PrioritizeVerb/",
							BindVerb:       "BindVerb/",
							Weight:         1,
						},
					},
				},
				port: 80,
			},
			want: func() *config.KubeSchedulerConfiguration {
				cfg := configGeneratedFromDefault()
				cfg.Profiles[0].Plugins.Score.Enabled = []config.Plugin{
					{Name: "NodeResourcesBalancedAllocationWrapped", Weight: 1},
					{Name: "InterPodAffinityWrapped", Weight: 1},
					{Name: "NodeAffinityWrapped", Weight: 1},
					{Name: "PodTopologySpreadWrapped", Weight: 2},
					{Name: "TaintTolerationWrapped", Weight: 1},
				}
				return &cfg
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := convertConfigurationForSimulator(tt.args.versioned, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertConfigurationForSimulator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Profiles) != len(tt.want.Profiles) {
				t.Errorf("unmatch length of profiles, want: %v, got: %v", len(tt.want.Profiles), len(got.Profiles))
				return
			}

			for k := range got.Profiles {
				sort.SliceStable(got.Profiles[k].PluginConfig, func(i, j int) bool {
					return got.Profiles[k].PluginConfig[i].Name < got.Profiles[k].PluginConfig[j].Name
				})
				sort.SliceStable(tt.want.Profiles[k].PluginConfig, func(i, j int) bool {
					return tt.want.Profiles[k].PluginConfig[i].Name < tt.want.Profiles[k].PluginConfig[j].Name
				})
			}

			assert.Equal(t, tt.want.Profiles[0].Plugins, got.Profiles[0].Plugins)
		})
	}
}

func configGeneratedFromDefault() config.KubeSchedulerConfiguration {
	var weight1 int32 = 1
	var weight2 int32 = 2
	versioned, _ := schedConfig.DefaultSchedulerConfig()
	cfg := versioned.DeepCopy()
	cfg.Profiles[0].Plugins.Bind.Enabled = []v1beta2config.Plugin{
		{Name: "DefaultBinderWrapped"},
	}
	cfg.Profiles[0].Plugins.PreFilter.Enabled = []v1beta2config.Plugin{
		{Name: "NodeResourcesFitWrapped"},
		{Name: "NodePortsWrapped"},
		{Name: "VolumeRestrictionsWrapped"},
		{Name: "PodTopologySpreadWrapped"},
		{Name: "InterPodAffinityWrapped"},
		{Name: "VolumeBindingWrapped"},
		{Name: "NodeAffinityWrapped"},
	}
	cfg.Profiles[0].Plugins.Filter.Enabled = []v1beta2config.Plugin{
		{Name: "NodeUnschedulableWrapped"},
		{Name: "NodeNameWrapped"},
		{Name: "TaintTolerationWrapped"},
		{Name: "NodeAffinityWrapped"},
		{Name: "NodePortsWrapped"},
		{Name: "NodeResourcesFitWrapped"},
		{Name: "VolumeRestrictionsWrapped"},
		{Name: "EBSLimitsWrapped"},
		{Name: "GCEPDLimitsWrapped"},
		{Name: "NodeVolumeLimitsWrapped"},
		{Name: "AzureDiskLimitsWrapped"},
		{Name: "VolumeBindingWrapped"},
		{Name: "VolumeZoneWrapped"},
		{Name: "PodTopologySpreadWrapped"},
		{Name: "InterPodAffinityWrapped"},
	}
	cfg.Profiles[0].Plugins.PostFilter.Enabled = []v1beta2config.Plugin{
		{Name: "DefaultPreemptionWrapped"},
	}
	cfg.Profiles[0].Plugins.Reserve.Enabled = []v1beta2config.Plugin{
		{Name: "VolumeBindingWrapped"},
	}
	cfg.Profiles[0].Plugins.PreBind.Enabled = []v1beta2config.Plugin{
		{Name: "VolumeBindingWrapped"},
	}
	cfg.Profiles[0].Plugins.PreScore.Enabled = []v1beta2config.Plugin{
		{Name: "InterPodAffinityWrapped"},
		{Name: "PodTopologySpreadWrapped"},
		{Name: "TaintTolerationWrapped"},
		{Name: "NodeAffinityWrapped"},
	}
	cfg.Profiles[0].Plugins.Score.Enabled = []v1beta2config.Plugin{
		{Name: "NodeResourcesBalancedAllocationWrapped", Weight: &weight1},
		{Name: "ImageLocalityWrapped", Weight: &weight1},
		{Name: "InterPodAffinityWrapped", Weight: &weight1},
		{Name: "NodeResourcesFitWrapped", Weight: &weight1},
		{Name: "NodeAffinityWrapped", Weight: &weight1},
		{Name: "PodTopologySpreadWrapped", Weight: &weight2},
		{Name: "TaintTolerationWrapped", Weight: &weight1},
	}
	pcMap := map[string]runtime.RawExtension{}
	for _, c := range cfg.Profiles[0].PluginConfig {
		pcMap[c.Name] = c.Args
	}

	var newpc []v1beta2config.PluginConfig
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeResourcesBalancedAllocationWrapped",
		Args: pcMap["NodeResourcesBalancedAllocation"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "InterPodAffinityWrapped",
		Args: pcMap["InterPodAffinity"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeResourcesFitWrapped",
		Args: pcMap["NodeResourcesFit"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "NodeAffinityWrapped",
		Args: pcMap["NodeAffinity"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "PodTopologySpreadWrapped",
		Args: pcMap["PodTopologySpread"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "VolumeBindingWrapped",
		Args: pcMap["VolumeBinding"],
	})
	newpc = append(newpc, v1beta2config.PluginConfig{
		Name: "DefaultPreemptionWrapped",
		Args: pcMap["DefaultPreemption"],
	})

	cfg.Profiles[0].PluginConfig = append(cfg.Profiles[0].PluginConfig, newpc...)

	converted := config.KubeSchedulerConfiguration{}
	scheme.Scheme.Convert(cfg, &converted, nil)
	converted.SetGroupVersionKind(v1beta2config.SchemeGroupVersion.WithKind("KubeSchedulerConfiguration"))

	converted.Extenders = []config.Extender{
		{
			URLPrefix:      "http://localhost:80/api/v1/extender/",
			PreemptVerb:    "preempt/0",
			FilterVerb:     "filter/0",
			PrioritizeVerb: "prioritize/0",
			BindVerb:       "bind/0",
			Weight:         1,
		},
	}
	return converted
}
