# Tailwind is pinned to v3: the project uses tailwind.config.js and v4 produces broken CSS
TAILWIND_VERSION := v3.4.17

buildall: templ build

build:
	go build -o tmp/main .

test:
	go test -v ./... -count=1 


tailwind:
	tailwindcss --config tailwind.config.js -i styles/input.css -o static/css/styles.css -m

templ:
	templ generate

install-tools:
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-linux-x64
	sudo chmod +x tailwindcss-linux-x64
	sudo mv tailwindcss-linux-x64 /usr/bin/tailwindcss
