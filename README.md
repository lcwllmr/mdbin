# mdbin

Simple pastebin-like client and server for quickly sharing Markdown content containing math.

## Installation

Currently there is only one way of installing the program: `go install github.com/lcwllmr/mdbin`.
If only the server is needed, you can also use the Docker image from DockerHub using `docker pull lcwllmr/mdbin:latest` (currently supports `amd64` and `arm64`).

Soon, I will also provide a Nix package that you can install via the flake.

## Usage

The executable `mdbin` acts both as server and client using sub-commands.

### Server-side

To launch the server, run
```bash
mdbin serve -port 23342 -htmldir ./html
```
The shown port is the default one, so leaving it out will use it either way.
Leaving out the `--htmldir` option will default to a fresh temporary directory, so you can't expect persistence between restarts.

The Docker container is launched by simply binding the port and the internal persistent `/html` directory. E.g.
```bash
docker run -d -p 23342:23342 -v $(PWD)/mdbin-html-store lcwllmr/mdbin
```

### Client-side

Say, you're dealing with the `./README.md`.
To quickly preview how it would look when uploaded, run
```bash
mdbin preview -file ./README.md
```
and open the displayed link in your browser.
If you leave this local server running, any changes will be detected and re-rendered automatically using web sockets.

As soon as you are ready to upload, run
```bash
mdbin push -server http://localhost:23342 -file ./README.md
```
Make sure to point this to the right server.
It will print a URL at which the content can be accessed.
