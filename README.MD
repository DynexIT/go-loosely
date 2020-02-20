# GO (Loosely)

Can work like a normal go compiler, but will try to strip modules of their names, so that (at least in our case) Go Plugins that include some package `x` , do not conflict when opened by an app that includes the same package `x`. This was used to compile a set of plugins we are using, the provided integration layers between other go programs.

## Installation

Quite simply, just run `go install` and make sure that the Go bin folder is included in the path.

## Usage

In order to compile a plugin, you can use the normal command to build, but by replacing the command to `go-loosely`:

`go-loosely build -o someOutputFile.so --buildmode=plugin`

All arguments will be passed to the go compiler, except if they have the "loose" argument prefix: `~`

The loose part of the compiler accepts the following parameters:

* `copy` - Specifies a file to copy from the build (as your project will be cloned to another area, so any output files are lost after deletion). This will have to be specified every time.
* `module` - Specifies the name of the module, in order for a rename to happen, to satisfy the requirements for Go to build it.
* `moveVendor` - After Go-Loosely changes to a vendoring mode, this can be used to strip out the common packages, that will be used by both the plugin, and the executable that uses the plugin. This parameter should be in the format `existingPackageName:newPackageName`

Example: 

Assuming a module is hosted on `github.com/joe/someModuleFolder`

```shell script
someModuleFolder> go-loosely build -o plugin.so -tags some_tag --buildmode=plugin ~copy=plugin.so ~module=github.com/joe/someModuleFolder ~moveVendor=github.com/joe/sharedPluginPackage:github.com/joe/someUniqueName
```

It should be noted that the new unique package name used can be anything - it doesn't really care too much, as long as it builds. Should also be noted that this is tied heavily with go modules.