package htmlform

var FormTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8"/>
	<title>{{.Title}}</title>
	<style>
		* { font-family: "Lucida Console", Monaco, monospace; }
		label { display: block; }
		form { margin: 10px; }
		form div { padding-bottom: 10px; }
	</style>
</head>
<body>
<h1>{{.Title}}</h1>
<form action="." method="post" enctype="multipart/form-data">
	{{range .Fields}}
		<div>
			{{if eq .Type "checkbox"}}
				<input type="checkbox" id="{{.Name}}" name="{{.Name}}" value="true" {{if eq .Value "true"}}checked{{end}}/>
				<label style="display: inline" for="{{.Name}}">{{.Label}}</label>
			{{else if eq .Type "select"}}
				<label for="{{.Name}}">{{.Label}}:</label>
				<select id="{{.Name}}" name="{{.Name}}" required>
					{{$selectValue := .Value}}
					{{range .Options}}
						<option value="{{.Value}}" {{if eq (printf "%v" .Value) $selectValue}}selected{{end}}>{{.Label}}</option>
					{{end}}
				</select>
			{{else}}
				<label for="{{.Name}}">{{.Label}}:</label>
				<input type="{{.Type}}" id="{{.Name}}" name="{{.Name}}" value="{{.Value}}" size="40" required/>
			{{end}}
		</div>
	{{end}}
	<button>{{.SubmitButtonText}}</button>
</form>
`
