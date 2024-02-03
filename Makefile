build:
	go build -o tmp/main .

buildall: tailwindcss templ build

test:
	go test -v ./... -count=1 


tailwind:
	tailwindcss --config tailwindcss/tailwind.config.js -i tailwindcss/input.css -o static/css/styles.css -m

templ:
	templ generate

