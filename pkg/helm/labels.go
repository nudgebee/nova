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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// ReleaseSourceSecrets reads releases from Helm's release Secrets (needs
	// cluster-wide list on secrets).
	ReleaseSourceSecrets = "secrets"
	// ReleaseSourceLabels derives releases from the labels and annotations Helm
	// puts on the objects it installs. Needs no Secrets access.
	ReleaseSourceLabels = "labels"

	chartLabel                 = "helm.sh/chart"
	legacyChartLabel           = "chart" // older charts (e.g. some prometheus-community ones) set only this
	appVersionLabel            = "app.kubernetes.io/version"
	releaseNameAnnotation      = "meta.helm.sh/release-name"
	releaseNamespaceAnnotation = "meta.helm.sh/release-namespace"

	labelListPageSize = 500

	// releaseSourceKey marks, in release.Labels, where nova found a release,
	// so Artifact Hub matching can account for the missing chart metadata.
	releaseSourceKey = "nova.release-source"
)

// labelledObject is the metadata of one object that carries a helm.sh/chart label.
type labelledObject struct {
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// objectLister lists the metadata of one kind of object, one page at a time.
type objectLister struct {
	kind string
	list func(ctx context.Context, namespace string, opts metav1.ListOptions) ([]metav1.ObjectMeta, string, error)
}

func labelListers(client kubernetes.Interface) []objectLister {
	return []objectLister{
		{"deployments", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.AppsV1().Deployments(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
		{"statefulsets", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.AppsV1().StatefulSets(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
		{"daemonsets", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.AppsV1().DaemonSets(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
		{"cronjobs", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.BatchV1().CronJobs(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
		{"services", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.CoreV1().Services(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
		{"configmaps", func(ctx context.Context, ns string, o metav1.ListOptions) ([]metav1.ObjectMeta, string, error) {
			l, err := client.CoreV1().ConfigMaps(ns).List(ctx, o)
			if err != nil {
				return nil, "", err
			}
			out := make([]metav1.ObjectMeta, len(l.Items))
			for i := range l.Items {
				out[i] = l.Items[i].ObjectMeta
			}
			return out, l.Continue, nil
		}},
	}
}

// GetHelmReleasesFromLabels derives the installed Helm releases from the
// helm.sh/chart label (or the older "chart" label) and the
// meta.helm.sh/release-* annotations on workloads,
// cronjobs, services and configmaps. It needs list on those kinds only, never
// on Secrets.
//
// What it cannot see compared to the Secrets source: releases that install
// none of those kinds (e.g. CRD-only charts), and chart metadata (home,
// sources, maintainers) used to score Artifact Hub matches. A release whose
// objects carry more than one chart label (an umbrella chart, where each
// subchart labels its own objects) is skipped: the parent chart is not
// recoverable from labels, and reporting a subchart would recommend the
// wrong upgrade.
func GetHelmReleasesFromLabels(ctx context.Context, client kubernetes.Interface, namespace string) ([]*release.Release, error) {
	var objects []labelledObject
	var errs []error
	listers := labelListers(client)
	selectors := []string{chartLabel, legacyChartLabel + ",!" + chartLabel}
	for _, l := range listers {
		for _, selector := range selectors {
			opts := metav1.ListOptions{LabelSelector: selector, Limit: labelListPageSize}
			for {
				items, cont, err := l.list(ctx, namespace, opts)
				if err != nil {
					klog.V(2).Infof("listing %s (%s) for helm labels: %v", l.kind, selector, err)
					errs = append(errs, fmt.Errorf("%s: %w", l.kind, err))
					break
				}
				for _, m := range items {
					objects = append(objects, labelledObject{Namespace: m.Namespace, Labels: m.Labels, Annotations: m.Annotations})
				}
				if cont == "" {
					break
				}
				opts.Continue = cont
			}
		}
	}
	if len(errs) == len(listers)*len(selectors) {
		return nil, fmt.Errorf("could not list any labelled objects: %w", errors.Join(errs...))
	}
	return releasesFromObjects(objects), nil
}

type releaseKey struct {
	namespace string
	name      string
}

type releaseLabels struct {
	charts      map[string]bool
	appVersions map[string]bool
}

// releasesFromObjects groups labelled objects by release and returns one
// release per group that names exactly one parseable chart.
func releasesFromObjects(objects []labelledObject) []*release.Release {
	groups := map[releaseKey]*releaseLabels{}
	for _, o := range objects {
		name := o.Annotations[releaseNameAnnotation]
		chartValue := o.Labels[chartLabel]
		if chartValue == "" {
			chartValue = o.Labels[legacyChartLabel]
		}
		if name == "" || chartValue == "" {
			continue
		}
		ns := o.Annotations[releaseNamespaceAnnotation]
		if ns == "" {
			ns = o.Namespace
		}
		k := releaseKey{namespace: ns, name: name}
		g := groups[k]
		if g == nil {
			g = &releaseLabels{charts: map[string]bool{}, appVersions: map[string]bool{}}
			groups[k] = g
		}
		g.charts[chartValue] = true
		if v := o.Labels[appVersionLabel]; v != "" {
			g.appVersions[v] = true
		}
	}

	keys := make([]releaseKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].namespace != keys[j].namespace {
			return keys[i].namespace < keys[j].namespace
		}
		return keys[i].name < keys[j].name
	})

	var out []*release.Release
	for _, k := range keys {
		g := groups[k]
		if len(g.charts) != 1 {
			klog.V(2).Infof("skipping release %s/%s: objects carry %d different chart labels", k.namespace, k.name, len(g.charts))
			continue
		}
		var chartValue string
		for c := range g.charts {
			chartValue = c
		}
		chartName, chartVersion, ok := splitChartLabel(chartValue)
		if !ok {
			klog.V(2).Infof("skipping release %s/%s: cannot split chart label %q into name and version", k.namespace, k.name, chartValue)
			continue
		}
		var appVersion string
		if len(g.appVersions) == 1 {
			for v := range g.appVersions {
				appVersion = v
			}
		}
		out = append(out, &release.Release{
			Name:      k.name,
			Namespace: k.namespace,
			Chart: &chart.Chart{Metadata: &chart.Metadata{
				Name:       chartName,
				Version:    chartVersion,
				AppVersion: appVersion,
			}},
			Info:   &release.Info{Status: release.StatusDeployed},
			Labels: map[string]string{releaseSourceKey: ReleaseSourceLabels},
		})
	}
	return out
}

// splitChartLabel splits a helm.sh/chart label value ("<name>-<version>",
// with "+" written as "_") into chart name and version. Both parts may
// contain "-", so it takes the leftmost "-" after which the rest is a strict
// semantic version (x.y.z). Failing that, it accepts a shorter "x.y" version
// (Helm tolerates those), still requiring a dot so a numeric name part such as
// "foo-2-bar" is not mistaken for a version.
func splitChartLabel(value string) (string, string, bool) {
	for _, strict := range []bool{true, false} {
		for i := 1; i < len(value); i++ {
			if value[i] != '-' {
				continue
			}
			candidate := strings.ReplaceAll(value[i+1:], "_", "+")
			bare := strings.TrimPrefix(candidate, "v")
			if strict {
				if _, err := semver.StrictNewVersion(bare); err == nil {
					return value[:i], candidate, true
				}
				continue
			}
			core, _, _ := strings.Cut(strings.SplitN(bare, "+", 2)[0], "-")
			if !strings.Contains(core, ".") {
				continue
			}
			if _, err := semver.NewVersion(bare); err == nil {
				return value[:i], candidate, true
			}
		}
	}
	return "", "", false
}
