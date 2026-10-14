package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
)

// SenseNova's Token Plan is a preset (#749): one key, its three APIs at
// token.sensenova.cn, its models.dev catalog, and an entry imported at its
// bases is the preset.
func TestSenseNovaPreset(t *testing.T) {
	p, err := FromPreset("sensenova")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "SenseNova" || p.Icon != "sensenova-color" || p.Catalog != "sensenova" ||
		p.Website != "https://platform.sensenova.cn" || p.KeysURL != "https://platform.sensenova.cn/console" {
		t.Fatalf("sensenova: %+v", p)
	}
	// Its Anthropic base is the host itself, so the route magpie asks is
	// /v1/messages; the vendor's /anthropic prefix answers 404.
	if p.Chat != "https://token.sensenova.cn/v1" || p.Responses != "https://token.sensenova.cn/v1" ||
		p.Anthropic != "https://token.sensenova.cn" {
		t.Fatalf("endpoints: %q %q %q", p.Chat, p.Responses, p.Anthropic)
	}
	if pr := Preset("sensenova"); pr.Kind != KindVendor || pr.NoList || pr.NoKey {
		t.Fatalf("preset: %+v", pr)
	}
	im, why := imported("default", "k", endpoints{chat: p.Chat + "/", responses: p.Responses + "/", anthropic: p.Anthropic}, nil)
	if why != "" || im.Preset != "sensenova" || im.Name != "SenseNova" || im.Icon != "sensenova-color" {
		t.Fatalf("imported: %+v (%s)", im, why)
	}
}

// Its /v1/models, as the vendor answers it: an OpenRouter-shaped list of the
// platform's models, in its order. Every one is kept — the image model
// sensenova-u1.5-lite and the deepseek-v4.1-flash this key's plan doesn't buy
// among them, since the list calls each of them "tokenplan" and only a
// request tells them apart (404 "model is not found" on the first, 403 "model
// is not available in the current token plan" on the second).
func TestSenseNovaLiveList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer k" {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		rw.Write([]byte(`{"data":[
{"id":"deepseek-v4-flash","name":"deepseek-v4-flash","created":1777392000,"input_modalities":["text"],"output_modalities":["text"],"context_length":1048576,"max_output_length":65536,"supported_features":["tools","json_mode","reasoning"],"businesses":["tokenplan","metered"]},
{"id":"sensenova-u1.5-lite","name":"sensenova-u1.5-lite","created":1777392000,"input_modalities":["text"],"output_modalities":["text"],"context_length":262144,"max_output_length":65536,"supported_features":["tools","json_mode","reasoning"],"businesses":["tokenplan","metered"]},
{"id":"deepseek-v4.1-flash","name":"deepseek-v4.1-flash","created":1777392000,"input_modalities":["text"],"output_modalities":["text"],"context_length":1048576,"max_output_length":65536,"supported_features":["tools","json_mode","reasoning"],"businesses":["tokenplan","metered"]}]}`))
	}))
	defer srv.Close()
	ms, _, err := catalog.FetchAt(context.Background(), srv.URL+"/v1", "k", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 3 || ms[0].ID != "deepseek-v4-flash" || ms[1].ID != "sensenova-u1.5-lite" || ms[2].ID != "deepseek-v4.1-flash" {
		t.Fatalf("%+v", ms)
	}
}
