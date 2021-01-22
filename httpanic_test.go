package httpanic

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBecause(t *testing.T) {
	testErr := errors.New("test error, please ignore")
	for tn, tc := range map[string]struct {
		err        error
		additional []Detail
		want       Reason
	}{
		"no additional reasons": {
			err: testErr,
			want: Reason{
				error:  testErr,
				Status: http.StatusInternalServerError,
			},
		},
		"with status code": {
			err: testErr,
			additional: []Detail{
				WithStatus(420),
			},
			want: Reason{
				error:  testErr,
				Status: 420,
			},
		},
		"with explanation": {
			err: testErr,
			additional: []Detail{
				WithExplanation("Chill, man!"),
			},
			want: Reason{
				error:       testErr,
				Status:      http.StatusInternalServerError,
				Explanation: "Chill, man!",
			},
		},
		"with status code and explanation": {
			err: testErr,
			additional: []Detail{
				WithStatus(420),
				WithExplanation("Chill, man!"),
			},
			want: Reason{
				error:       testErr,
				Status:      420,
				Explanation: "Chill, man!",
			},
		},
		"latest additional reason wins": {
			err: testErr,
			additional: []Detail{
				WithStatus(417),
				WithStatus(418),
				WithStatus(419),
				WithExplanation("Chill, man!"),
				WithStatus(420),
			},
			want: Reason{
				error:       testErr,
				Status:      420,
				Explanation: "Chill, man!",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			got := Because(tc.err, tc.additional...)
			if diff := cmp.Diff(tc.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Because() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestRecoverFromPanic(t *testing.T) {
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreUnexported(httptest.ResponseRecorder{}),
		cmpopts.IgnoreFields(httptest.ResponseRecorder{}, "HeaderMap"),
	}
	for tn, tc := range map[string]struct {
		p    interface{}
		want httptest.ResponseRecorder
	}{
		"nil": {
			want: httptest.ResponseRecorder{
				Code: http.StatusInternalServerError,
			},
		},
		"without reason": {
			p: errors.New("rut-ro raggy"),
			want: httptest.ResponseRecorder{
				Code: http.StatusInternalServerError,
			},
		},
		"default reason": {
			p: Because(errors.New("rut-ro raggy")),
			want: httptest.ResponseRecorder{
				Code: http.StatusInternalServerError,
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			got := httptest.ResponseRecorder{}
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("recoverFromPanic(): a panic escaped: %v", r)
				}
			}()
			func(t *testing.T, w http.ResponseWriter) {
				defer recoverFromPanic(w)
				panic(tc.p)
			}(t, &got)
			if diff := cmp.Diff(tc.want, got, cmpOpts...); diff != "" {
				t.Errorf("recoverFromPanic(): mismatch (-want, +got):\n%v", diff)
			}
		})
	}
}
