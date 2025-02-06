package main

import (
	"testing"

	"github.com/evbruno/go-pf/lib"
)

func TestGeneratePortForwardCommand(t *testing.T) {
	tests := []struct {
		name string
		svc  lib.K8SService
		want string
	}{
		{
			name: "Single Port",
			svc: lib.K8SService{
				Context:   "test-context",
				Namespace: "test-namespace",
				Name:      "test-service",
				Ports:     []string{"8080"},
			},
			want: "kubectl --context test-context -n test-namespace port-forward service/test-service 8080",
		},
		{
			name: "Multiple Ports",
			svc: lib.K8SService{
				Context:   "test-context",
				Namespace: "test-namespace",
				Name:      "test-service",
				Ports:     []string{"8080", "3000"},
			},
			want: "kubectl --context test-context -n test-namespace port-forward service/test-service 8080 3000",
		},
		{
			name: "No Ports",
			svc: lib.K8SService{
				Context:   "test-context",
				Namespace: "test-namespace",
				Name:      "test-service",
				Ports:     []string{},
			},
			want: "kubectl --context test-context -n test-namespace port-forward service/test-service ", // Note the trailing space
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lib.GeneratePortForwardCommand(tt.svc); got != tt.want {
				t.Errorf("GeneratePortForwardCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
