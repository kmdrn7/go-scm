// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
)

func TestWebhooks(t *testing.T) {
	tests := []struct {
		sig    string
		event  string
		before string
		after  string
		obj    interface{}
	}{
		//
		// push events
		//

		// push hooks
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},

		//
		// tag events
		//

		// create
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push_tag_create.json",
			after:  "testdata/webhooks/push_tag_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push_tag_delete.json",
			after:  "testdata/webhooks/push_tag_delete.json.golden",
			obj:    new(scm.TagHook),
		},

		//
		// branch events
		//

		// create
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push_branch_create.json",
			after:  "testdata/webhooks/push_branch_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// delete
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "repo:push",
			before: "testdata/webhooks/push_branch_delete.json",
			after:  "testdata/webhooks/push_branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},

		//
		// pull request events
		//

		// pull request created
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:created",
			before: "testdata/webhooks/pr_created.json",
			after:  "testdata/webhooks/pr_created.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request created, having branch name of pattern foo/bar
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:created",
			before: "testdata/webhooks/pr_created_slashbranch.json",
			after:  "testdata/webhooks/pr_created_slashbranch.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request updated
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:updated",
			before: "testdata/webhooks/pr_updated.json",
			after:  "testdata/webhooks/pr_updated.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request fulfilled (merged)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:fulfilled",
			before: "testdata/webhooks/pr_fulfilled.json",
			after:  "testdata/webhooks/pr_fulfilled.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request rejected (closed, declined)
		{
			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
			event:  "pullrequest:rejected",
			before: "testdata/webhooks/pr_declined.json",
			after:  "testdata/webhooks/pr_declined.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// 		// pull request labeled
		// 		{
		// 			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
		// 			event:  "pull_request",
		// 			before: "samples/pr_labeled.json",
		// 			after:  "samples/pr_labeled.json.golden",
		// 			obj:    new(scm.PullRequestHook),
		// 		},
		// 		// pull request unlabeled
		// 		{
		// 			sig:    "71295b197fa25f4356d2fb9965df3f2379d903d7",
		// 			event:  "pull_request",
		// 			before: "samples/pr_unlabeled.json",
		// 			after:  "samples/pr_unlabeled.json.golden",
		// 			obj:    new(scm.PullRequestHook),
		// 		},
	}

	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			before, err := os.ReadFile(test.before)
			if err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(test.after)
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("GET", "/?secret=71295b197fa25f4356d2fb9965df3f2379d903d7", buf)
			r.Header.Set("x-event-key", test.event)
			r.Header.Set("X-Hook-UUID", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

			s := new(webhookService)
			o, err := s.Parse(r, secretFunc)
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(after, &test.obj)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.obj, o); diff != "" {
				t.Errorf("Error unmarshaling %s", test.before)
				t.Log(diff)

				// debug only. remove once implemented
				err := json.NewEncoder(os.Stdout).Encode(o)
				if err != nil {
					t.Fatal(err)
				}
			}

			switch event := o.(type) {
			case *scm.PushHook:
				if !strings.HasPrefix(event.Ref, "refs/") {
					t.Errorf("Push hook reference must start with refs/")
				}
			case *scm.BranchHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			case *scm.TagHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			}
		})
	}
}

func TestWebhookInvalid(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		url     string
		headers map[string]string
	}{
		{
			name: "validate webhook via key from query param",
			body: "testdata/webhooks/push.json",
			url:  "/?secret=xxxxxinvalidxxxxxx",
			headers: map[string]string{
				"x-event-key": "repo:push",
			},
		},
		{
			name: "validate webhook via hmac signature from header",
			body: "testdata/webhooks/push.json",
			url:  "/",
			headers: map[string]string{
				"x-event-key":     "repo:push",
				"x-hub-signature": "sha256=xxxxxinvalidxxxxxx",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, _ := os.ReadFile(test.body)
			r, _ := http.NewRequest("GET", test.url, bytes.NewBuffer(f))
			for k, v := range test.headers {
				r.Header.Set(k, v)
			}

			s := new(webhookService)
			_, err := s.Parse(r, secretFunc)
			if err != scm.ErrSignatureInvalid {
				t.Errorf("Expect invalid signature error, got %v", err)
			}
		})
	}
}

func TestWebhookValidated(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		url     string
		headers map[string]string
	}{
		{
			name: "validate webhook via key from query param",
			body: "testdata/webhooks/push.json",
			url:  "/?secret=71295b197fa25f4356d2fb9965df3f2379d903d7",
			headers: map[string]string{
				"x-event-key": "repo:push",
			},
		},
		{
			name: "validate webhook via hmac signature from header",
			body: "testdata/webhooks/push.json",
			url:  "/",
			headers: map[string]string{
				"x-event-key":     "repo:push",
				"x-hub-signature": "sha256=811688563e3cc0d2bd3f5b277b0b5835c06cbc0bbefeb582498888925675d014",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, _ := os.ReadFile(test.body)
			r, _ := http.NewRequest("GET", test.url, bytes.NewBuffer(f))
			for k, v := range test.headers {
				r.Header.Set(k, v)
			}

			s := new(webhookService)
			_, err := s.Parse(r, secretFunc)
			if err != nil {
				t.Errorf("Expect valid signature, got %v", err)
			}
		})
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "71295b197fa25f4356d2fb9965df3f2379d903d7", nil
}
