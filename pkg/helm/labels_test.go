// Copyright 2026 Nudgebee, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestSplitChartLabel(t *testing.T) {
	tests := []struct {
		in, name, version string
		ok                bool
	}{
		{"cert-manager-v1.13.1", "cert-manager", "v1.13.1", true},
		{"my-agent-0.1.33-rc-4", "my-agent", "0.1.33-rc-4", true},
		{"redis-17.11.8", "redis", "17.11.8", true},
		{"app-0.1", "app", "0.1", true},
		{"foo-2-bar-1.0.0", "foo-2-bar", "1.0.0", true},
		{"foo-2-bar-0.3", "foo-2-bar", "0.3", true},
		{"chart-1.2.3_build.5", "chart", "1.2.3+build.5", true},
		{"no-version", "", "", false},
		{"chart-2", "", "", false},
		{"", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			name, version, ok := splitChartLabel(tt.in)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.version, version)
		})
	}
}

func obj(ns, release, chartLabelValue string) labelledObject {
	return labelledObject{
		Namespace:   ns,
		Labels:      map[string]string{chartLabel: chartLabelValue},
		Annotations: map[string]string{releaseNameAnnotation: release, releaseNamespaceAnnotation: ns},
	}
}

func TestReleasesFromObjects(t *testing.T) {
	objects := []labelledObject{
		obj("web", "ingress", "ingress-nginx-4.10.1"),
		obj("web", "ingress", "ingress-nginx-4.10.1"), // deployment + service of one release
		// umbrella chart: each subchart labels its own objects
		obj("mon", "stack", "kube-state-metrics-5.28.1"),
		obj("mon", "stack", "prometheus-node-exporter-4.43.1"),
		// same release name in another namespace is a different release
		obj("web2", "ingress", "ingress-nginx-4.9.0"),
		// no release annotation (not installed by Helm 3) is ignored
		{Namespace: "x", Labels: map[string]string{chartLabel: "thing-1.0.0"}},
		// unparseable chart label is skipped
		obj("y", "odd", "odd"),
	}
	objects[0].Labels[appVersionLabel] = "1.10.1"

	got := releasesFromObjects(objects)
	require.Len(t, got, 2)

	assert.Equal(t, "ingress", got[0].Name)
	assert.Equal(t, "web", got[0].Namespace)
	assert.Equal(t, "ingress-nginx", got[0].Chart.Metadata.Name)
	assert.Equal(t, "4.10.1", got[0].Chart.Metadata.Version)
	assert.Equal(t, "1.10.1", got[0].Chart.Metadata.AppVersion)
	require.NotNil(t, got[0].Info)
	assert.Equal(t, ReleaseSourceLabels, got[0].Labels[releaseSourceKey])

	assert.Equal(t, "web2", got[1].Namespace)
	assert.Equal(t, "4.9.0", got[1].Chart.Metadata.Version)
}

func TestReleasesFromObjects_ReleaseNamespaceFallsBackToObjectNamespace(t *testing.T) {
	o := labelledObject{
		Namespace:   "apps",
		Labels:      map[string]string{chartLabel: "web-1.0.0"},
		Annotations: map[string]string{releaseNameAnnotation: "web"},
	}
	got := releasesFromObjects([]labelledObject{o})
	require.Len(t, got, 1)
	assert.Equal(t, "apps", got[0].Namespace)
}

func helmMeta(name, ns, release, chartLabelValue string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   ns,
		Labels:      map[string]string{chartLabel: chartLabelValue},
		Annotations: map[string]string{releaseNameAnnotation: release, releaseNamespaceAnnotation: ns},
	}
}

func TestGetHelmReleasesFromLabels_ForbiddenKindIsSkipped(t *testing.T) {
	client := fake.NewClientset(
		&appsv1.Deployment{ObjectMeta: helmMeta("api", "app", "api", "api-2.0.0")},
		&corev1.Service{ObjectMeta: helmMeta("db", "data", "db", "postgresql-12.1.0")},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "plain", Namespace: "data"}},
	)
	client.PrependReactor("list", "configmaps", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "configmaps"}, "", nil)
	})

	legacy := helmMeta("exporter", "mon", "exporter", "")
	legacy.Labels = map[string]string{legacyChartLabel: "prometheus-postgres-exporter-6.8.1"}
	require.NoError(t, client.Tracker().Add(&appsv1.Deployment{ObjectMeta: legacy}))

	got, err := GetHelmReleasesFromLabels(context.Background(), client, "")
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, "api", got[0].Chart.Metadata.Name)
	assert.Equal(t, "postgresql", got[1].Chart.Metadata.Name)
	assert.Equal(t, "prometheus-postgres-exporter", got[2].Chart.Metadata.Name)
	assert.Equal(t, "6.8.1", got[2].Chart.Metadata.Version)
}

func TestGetHelmReleasesFromLabels_AllKindsForbidden(t *testing.T) {
	client := fake.NewClientset()
	client.PrependReactor("list", "*", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{}, "", nil)
	})
	_, err := GetHelmReleasesFromLabels(context.Background(), client, "")
	assert.Error(t, err)
}

func TestFindBestArtifactHubMatch_LabelsRequireInstalledVersion(t *testing.T) {
	rel := releasesFromObjects([]labelledObject{obj("mon", "grafana", "grafana-11.1.1")})[0]
	popular := ArtifactHubHelmPackage{
		Name: "grafana", Version: "10.5.15", Stars: 500, Official: true,
		Repository:        ArtifactHubRepository{Name: "popular", VerifiedPublisher: true},
		AvailableVersions: []AvailableVersion{{Version: "10.5.15"}},
	}
	upstream := ArtifactHubHelmPackage{
		Name: "grafana", Version: "12.1.8", Stars: 1,
		Repository:        ArtifactHubRepository{Name: "upstream"},
		AvailableVersions: []AvailableVersion{{Version: "11.1.1"}, {Version: "12.1.8"}},
	}

	o := FindBestArtifactHubMatch(rel, []ArtifactHubHelmPackage{popular, upstream})
	require.NotNil(t, o)
	assert.Equal(t, "12.1.8", o.Latest.Version)

	o = FindBestArtifactHubMatch(rel, []ArtifactHubHelmPackage{popular})
	require.NotNil(t, o)
	assert.Empty(t, o.Latest.Version, "no package published 11.1.1, so no recommendation")
	assert.False(t, o.IsOld)
}

func TestPackagePublishedVersion_IgnoresVPrefix(t *testing.T) {
	pkg := ArtifactHubHelmPackage{AvailableVersions: []AvailableVersion{{Version: "1.13.1"}}}
	assert.True(t, packagePublishedVersion("v1.13.1", pkg))
	assert.True(t, packagePublishedVersion("1.13.1", pkg))
	assert.False(t, packagePublishedVersion("1.13.2", pkg))
}
