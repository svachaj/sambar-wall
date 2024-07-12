# Web application: Sambar Lezecká Stěna - Kroužky a registrace

## Prerequisites

### [Go](https://go.dev/) of course 🚀

See go mod file for the version of Go used in this project.

### Makefile command to install all the other tools needed for the project

`make install-tools`

### [TEMPL](https://templ.guide/) - Template Engine

To have the best templ experience, you should install VSCode extension `templ-vscode`.

Than follow this link how to setup IDE [IDE Support](https://templ.guide/commands-and-tools/ide-support)

### [Tailwindcss](https://tailwindcss.com/) - CSS Framework

With no need to install npm, you can use the standalone CLI to compile your Tailwind CSS files.
`https://tailwindcss.com/blog/standalone-cli`

To have the best tailwindcss experience, you should install VSCode extension `Tailwind CSS IntelliSense` and let the extension to know where the tailwindcss is located, because we are using .templ files instead of .html files.

Extension settings:
`"tailwindCSS.includeLanguages": {
  "templ": "html"
}`

### [Air](https://github.com/air-verse/air) - Live reload for Go apps

## How to develop

### 1. Clone the repository

### 2. Set up prerequisites

### 3. Create .env file in the root of the project to set up the environment variables for the configuration

`APP_PORT=5500`

### 3. Run with Air

`air`
