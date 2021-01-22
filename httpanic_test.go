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
				t.Errorf("Because(): return value mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

var errForTesting = errors.New("rut-ro raggy")

// cuzTest is a reasoner which creates does nothing fancy.
func cuzTest(e error, _ ...Detail) Reason {
	return Reason{error: e}
}

func TestRecoverFromPanic(t *testing.T) {
	cmpOpts := []cmp.Option{
		cmp.Comparer(func(x, y error) bool {
			// Compare the errors by value only.
			return x.Error() == y.Error()
		}),
	}

	for tn, tc := range map[string]struct {
		p           interface{}
		want        Reason
		shouldPanic bool
	}{
		"nil": {
			shouldPanic: true,
		},
		"arbitrary other non-reason": {
			p:           struct{ string }{"this would be weird, but might as well test for it"},
			shouldPanic: true,
		},
		"string": {
			p:    "this is a string",
			want: Reason{error: errors.New("this is a string")},
		},
		"error": {
			p:    errForTesting,
			want: Reason{error: errForTesting},
		},
		"reason": {
			p:    Because(errForTesting),
			want: Reason{error: errForTesting},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("recoverFromPanic(): unexpected panic: %v", r)
				}
			}()
			func(t *testing.T) {
				tcRender := func(w http.ResponseWriter, got Reason) {
					if diff := cmp.Diff(tc.want, got, cmpOpts...); diff != "" {
						t.Errorf("recoverFromPanic(): render argument mismatch (-want, +got):\n%v", diff)
					}
				}
				defer recoverFromPanic(&httptest.ResponseRecorder{}, tcRender, cuzTest)
				panic(tc.p)
			}(t)
		})
	}
}
