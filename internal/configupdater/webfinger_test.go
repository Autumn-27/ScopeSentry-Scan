package configupdater

import (
	"testing"

	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
)

func TestBuildACMatcher_FingerprintWithRequestConditionFallsBackToNonAC(t *testing.T) {
	fp := &types.Fingerprint{
		ID:   "alibaba-nacos",
		Name: "Alibaba-Nacos",
		Rules: []types.Rule{
			{
				Logic: "OR",
				Conditions: []types.Condition{
					{
						Location:  "request",
						MatchType: "active",
						Path:      "/nacos/",
					},
					{
						Location:  "body",
						MatchType: "contains",
						Pattern:   "console-ui/public/img/nacos-logo.png",
					},
					{
						Location:  "title",
						MatchType: "contains",
						Pattern:   "Nacos",
					},
				},
			},
		},
	}

	matcher := BuildACMatcher([]*types.Fingerprint{fp})

	if len(matcher.NonACFingerprints) != 1 || matcher.NonACFingerprints[0].ID != fp.ID {
		t.Fatalf("expected fingerprint with request condition to be in NonACFingerprints")
	}

	if len(matcher.TitlePatterns) != 0 || len(matcher.HeaderPatterns) != 0 || len(matcher.BodyPatterns) != 0 {
		t.Fatalf("expected no AC patterns for fingerprint containing request condition")
	}
}
