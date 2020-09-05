/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package matcher

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/onsi/gomega/types"
)

func HaveHalted() types.GomegaMatcher {
	return &statusMatcher{
		expected: api.DatabasePhaseHalted,
	}
}

type statusMatcher struct {
	expected api.DatabasePhase
}

func (matcher *statusMatcher) Match(actual interface{}) (success bool, err error) {
	phase := actual.(api.DatabasePhase)
	return phase == matcher.expected, nil
}

func (matcher *statusMatcher) FailureMessage(actual interface{}) (message string) {
	return "Expected to be Running all Pods"
}

func (matcher *statusMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return "Expected to be not Running all Pods"
}
