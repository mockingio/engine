package main

import (
    "dagger.io/dagger"

		"universe.dagger.io/bash"
		"universe.dagger.io/alpine"
    "universe.dagger.io/go"

		"github.com/mockingio/dagger/ci/golangci"
)

dagger.#Plan & {
    client: filesystem: ".": read: contents: dagger.#FS

    actions: {
				_source: client.filesystem.".".read.contents

				test: {
					unit: go.#Test & {
						source:  _source
						package: "./..."
						command: flags: "-race": true
					}
				}

				lint: {
						go: golangci.#Lint & {
						source:  _source
						version: "1.45"
					}
				}

				version: {
					_image: alpine.#Build & {
						packages: bash: _
						packages: curl: _
						packages: git: _
					}

					_revision: bash.#Run & {
						input:   _image.output
						workdir: "/src"
						mounts: source: {
							dest:     "/src"
							contents: _source
						}

						script: contents: #"""
							printf "$(git rev-parse --short HEAD)" > /revision
							"""#
						export: files: "/revision": string
					}

					output: _revision.export.files["/revision"]
				}
    }

}