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

				lint: {
						go: golangci.#Lint & {
						source:  _source
						version: "1.45"
					}
				}
    }

}