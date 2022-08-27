// Copyright 2022 CUE Authors
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

// serverless-local emulates Netlify serverless function deployment on a local
// development machine, serving functions from:
//
//     http://localhost:8081/.netlify/functions/$function
//
package main

import (
	"log"
	"net/http"

	"github.com/cue-lang/cuelang.org/internal/functions/gerritstatusupdater"
	"github.com/cue-lang/cuelang.org/internal/functions/snippets"
)

func main() {
	http.HandleFunc("/.netlify/functions/snippets", snippets.Function{DevelopmentMode: true}.ServeHTTP)
	http.HandleFunc("/.netlify/functions/gerritstatusupdater", gerritstatusupdater.Function{DevelopmentMode: true}.ServeHTTP)

	// In development mode, the playground TypeScript application knows
	// to query localhost:8081 for the /.netlify/functions/snippets endpoint
	log.Fatal(http.ListenAndServe(":8081", nil))
}