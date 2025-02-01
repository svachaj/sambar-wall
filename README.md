# Web application: Sambar Lezecká Stěna - Kroužky a registrace

https://registrace.stenakladno.cz/

## Important note

This application uses MSSQL as its database because it is a replacement for an existing interface to an older version of the application and is therefore tightly bound to MSSQL. Based on this application, a general template will be created, which should include an ORM to make everything more flexible.

The application is currently strictly localized to the Czech language, which is another aspect to be improved in the general template.

### The result should be a preconfigured template for developing modern and high-performance web applications based on Go and HTMX, with various authentication options.

## Auth

This application uses authentication via a so-called magic email link. Nowadays, authentication with just a username and password is not very secure. This type of authentication provides a higher level of security while also being extremely user-friendly, which is a crucial advantage.

Email clients today are inherently secure enough, so shifting the responsibility to them makes sense. While this method is not strictly two-factor authentication in the traditional sense, it certainly offers a much stronger security model than a simple username and password.

I highly recommend this authentication approach, and this repository can serve as a full reference for implementation and inspiration.

[Security handler](modules/security/security-handlers.go)

## Prerequisites

### [Go](https://go.dev/) of course 🚀

See go mod file for the version of Go used in this project.

### Makefile command to install all the other tools needed for the project

`make install-tools` (for Linux x64)

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
