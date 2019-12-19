## Contributing to Eunomia

The Eunomia project is licensed under [Apache 2.0 license](LICENSE) and is open for contributions via GitHub pull requests. 

### How to Contribute

You can contribute to this project by submitting pull requests, or by submitting issues for bugs/enhancemnts that you
found which will help us improve Eunomia.

When submitting a pull request, we expect that it will meet following requirements:

- Code must be written using [idiomatic Go](https://golang.org/doc/effective_go.html).
- Code must be formatted using [gofmt](https://golang.org/cmd/gofmt).
- Code must not produce [go lint](https://github.com/golang/lint) errors or warnings.
- Code must not produce [go vet](https://golang.org/cmd/vet) errors or warnings.
- All code changes are covered by the suite of unit and e2e tests.

### Commit Message Format

Every commit message should contain information about Go package under which the changes were applied, a short summary of the implemented changes and the GitHub issue it relates to (if applicable).

```
<package>: <short commit message>

Detailed commit message explaining why the code is being changed.

Fixes #<issue-number>
```

Example commit message.
```
main: add HTTP timeout to GitHub webhook

Prior to this change the net/http webserver used by the GitHub webhook
did not enforce an HTTP tiemout. Misbehaving HTTP clients could hold open
HTTP connections indefinitely.

Fixes #99999
```

Code changes that do not touch any Go source files should follow the below guidelines. These
code changes fall into two different categories "chores" and "real code". Chores are changes
to files like the Makefile, .travis.yml, or any markdown file documentation updates. Real code
changes are things like updating template processor images, but do not include any Go source file
changes.

Example "real code" changes commit message.
```
template/jinja: pin jinja version in template processor image

This change pins the jinja python module version in the jinja template
processor image to esnure a consitent version of jinja is installed. Prior
to this change the latest version of the jinja pip module would always be
install which could lead to surprise breakages everytime new template
processor images were built.

Fixes #99999
```

Example "chore" commit message.
```
chore: fail CI tests when golint errors are found

Adding the -set_exit_status CLI option to the golint command ensures that
the CI tests will fail when golint errors are found. Prior to this change
Travis CI builds would succeed when golint errors were found.

Fixes #99999
```