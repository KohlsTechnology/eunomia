 Contributing to Eunomia

Eunomia project is licensed under [Apache 2.0 license](LICENSE) and is open for contributions mainly via GitHub pull requests. 

## How to Contribute

You can contribute to the project by submitting Pull Requests (PRs), submitting Issues that you found which will help us improve Eunomia, or by suggesting new features.

When submitting a Pull Request, we expect that it will pass following requirements:

- Your code must be written in an idiomatic Go.
- Formatted in accordance with the [gofmt](https://golang.org/cmd/gofmt).
- [go lint](https://github.com/golang/lint) shouldn't produce any warnings, same as [go vet](https://golang.org/cmd/vet)
- If you are submitting PR with a new feature, code should be covered with the suite of unit tests that test new functionality. Same rule applies for PRs that are bug fixes.

### Commit message format

Every commit message should contain information about package under which the changes were applied, short summary of the implemented changes and GitHub issue it relates to (if applicable):

```
<package>: <what has changed> [Fixes #<issue-number>]
```
