/*
goos: darwin
goarch: arm64
pkg: k8s.io/kubernetes/pkg/kubelet/prober
cpu: Apple M2
                        │ nocaching.txt │    caching.txt     │
                        │    sec/op     │   sec/op     vs base   │
WorkerProbe_NoCaching-8     21.17m ± 3%
WorkerProbe_Caching-8                     250.8µ ± 2%
geomean                     21.17m        250.8µ       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean

                        │ nocaching.txt │     caching.txt     │
                        │     B/op      │     B/op      vs base   │
WorkerProbe_NoCaching-8    74.69Mi ± 0%
WorkerProbe_Caching-8                     385.8Ki ± 0%
geomean                    74.69Mi        385.8Ki       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean

                        │ nocaching.txt │    caching.txt     │
                        │   allocs/op   │  allocs/op   vs base   │
WorkerProbe_NoCaching-8     859.1k ± 0%
WorkerProbe_Caching-8                     3.634k ± 0%
geomean                     859.1k        3.634k       ? ¹ ²
¹ benchmark set differs from baseline; geomeans may not be comparable
² ratios must be >0 to compute geomean
*/

package prober

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const nPods = 110
const epoch = 5 * 60

func fakeWorkerDoProbe(container *v1.Container, holder *httpProbeRequestHolder, useCache bool) {
	for range epoch {
		if !useCache {
			holder.reset()
		}
		_, _ = holder.getRequest(container, "10.0.0.1")
		// perform the probe
	}
}

func runParallelWorkerProbes(useCache bool) {
	var containers = make([]v1.Container, nPods*2)
	var holders = make([]*httpProbeRequestHolder, nPods*2)
	for i := range nPods {
		containers[2*i] = v1.Container{Name: fmt.Sprintf("pod-%d-container-1", i)}
		containers[2*i+1] = v1.Container{Name: fmt.Sprintf("pod-%d-container-2", i)}
		holders[2*i] = &httpProbeRequestHolder{
			httpGet: &v1.HTTPGetAction{Port: intstr.FromInt(80)},
			podIP:   "10.0.0.1",
		}
		holders[2*i+1] = &httpProbeRequestHolder{
			httpGet: &v1.HTTPGetAction{Port: intstr.FromInt(80)},
			podIP:   "10.0.0.1",
		}
	}
	waiter := make(chan struct{})
	for i := range nPods {
		c1 := &containers[2*i]
		c2 := &containers[2*i+1]
		h1 := holders[2*i]
		h2 := holders[2*i+1]
		go func() {
			fakeWorkerDoProbe(c1, h1, useCache)
			fakeWorkerDoProbe(c2, h2, useCache)
			waiter <- struct{}{}
		}()
	}
	for range nPods {
		<-waiter
	}
}

func BenchmarkWorkerProbe_Caching(b *testing.B) {
	for b.Loop() {
		runParallelWorkerProbes(true) // use cache
	}
}

func BenchmarkWorkerProbe_NoCaching(b *testing.B) {
	for b.Loop() {
		runParallelWorkerProbes(false) // always reset, no cache
	}
}
