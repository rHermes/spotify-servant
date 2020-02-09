package main

import (
	"context"
	"html/template"
	"net/http"
)

const (
	IndexTemplateName        = "__ctx__tmpl_index"
	LoginTemplateName        = "__ctx__tmpl_login"
	ProfileIndexTemplateName = "__ctx__tmpl_profile_index"
)

// FirebaseConfig is needed for the firebase scripts
type FirebaseConfig struct {
	ApiKey            string
	AuthDomain        string
	DatabaseURL       string
	ProjectID         string
	StorageBucket     string
	MessagingSenderID string
	AppID             string
	MeasurementID     string
}

type TemplateBaseCtx struct {
	Firebase FirebaseConfig
}

// TemplateMiddleware creates the middlewares that we will use.
// Panics if the template is not good
func TemplateMiddleware(next http.Handler) http.Handler {
	firebaseTmpl := template.Must(template.New("base").Parse(FirebaseScriptSource))
	baseTmpl := firebaseTmpl
	layoutTmpl := template.Must(template.Must(baseTmpl.Clone()).Parse(LayoutTemplateSource))
	indexTmpl := template.Must(template.Must(layoutTmpl.Clone()).Parse(IndexTemplateSource))
	loginTmpl := template.Must(template.Must(layoutTmpl.Clone()).Parse(LoginTemplateSource))
	profileIndexTmpl := template.Must(template.Must(layoutTmpl.Clone()).Parse(ProfileIndexTemplateSource))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, IndexTemplateName, indexTmpl)
		ctx = context.WithValue(ctx, LoginTemplateName, loginTmpl)
		ctx = context.WithValue(ctx, ProfileIndexTemplateName, profileIndexTmpl)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTemplateFromCtx gets a template from the context
func GetTemplateFromCtx(ctx context.Context, name string) *template.Template {
	tmpl, ok := ctx.Value(name).(*template.Template)
	if !ok {
		return nil
	}

	return tmpl
}

// language=gohtml
const LayoutTemplateSource = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Spotify Servant</title>
    
    
    {{ block "stylesheet-include" .}}{{end}}
</head>
<body>
<h1>Spotify servant</h1>

{{ block "root" .}}{{end}}


<p>Powered by go and happiness</p>

<!-- The core Firebase JS SDK is always required and must be listed first -->
<script src="https://www.gstatic.com/firebasejs/7.8.0/firebase-app.js"></script>
<script src="https://www.gstatic.com/firebasejs/7.8.0/firebase-analytics.js"></script>
{{template "firebase-init" .Firebase}}
{{ block "scripts-include" .}}{{end}}
</body>
</html>`

// language=gohtml
const IndexTemplateSource = `{{define "root"}}
    <h2>
        This is the index template.
    </h2>


    <a href="/login">Log in</a>
    <a href="/logout">Log out</a>
    <a href="/profile">Profile</a>
{{end}}`

// language="gohtml"
const LoginTemplateSource = `{{define "stylesheet-include"}}
 <link type="text/css" rel="stylesheet" href="https://cdn.firebase.com/libs/firebaseui/3.5.2/firebaseui.css" />
{{end}}

{{define "root"}}
    <h2>
        This is the login template.
    </h2>
        
        <div id="firebaseui-auth-container"></div>
        
    
{{end}}

{{define "scripts-include"}}
    <script src="https://www.gstatic.com/firebasejs/7.8.0/firebase-auth.js"></script>
    <script src="https://cdn.firebase.com/libs/firebaseui/3.5.2/firebaseui.js"></script>

    <script src="/static/js/login.js"></script>
{{end}}`

// language=gohtml
const FirebaseScriptSource = `{{define "firebase-init"}}
<script>
  // Your web app's Firebase configuration
  var firebaseConfig = {
    apiKey: "{{.ApiKey}}",
    authDomain: "{{.AuthDomain}}",
    databaseURL: "{{.DatabaseURL}}",
    projectId: "{{.ProjectID}}",
    storageBucket: "{{.StorageBucket}}",
    messagingSenderId: "{{.MessagingSenderID}}",
    appId: "{{.AppID}}",
    measurementId: "{{.MeasurementID}}"
  };
  // Initialize Firebase
  firebase.initializeApp(firebaseConfig);
  firebase.analytics();
</script>
{{ end }}`

// language=gohtml
const ProfileIndexTemplateSource = `{{ define "root" }}
    <h2>This is your profile</h2>
{{ end }}`
