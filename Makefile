buildall: tailwind templ build

build:
	go build -o tmp/main .

test:
	go test -v ./... -count=1 


tailwind:
	tailwindcss --config tailwind.config.js -i styles/input.css -o static/css/styles.css -m

templ:
	templ generate

