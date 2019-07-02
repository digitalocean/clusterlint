# Contributing to clusterlint

We welcome all contributions to this project, whether in the form of pull
requests or issues. Please know that all contributions are appreciated, and we
will do our best to review code, answer questions, and resolve issues as time
allows.

That said, in order to maximize the effectiveness of your contributions, please
keep the following guidelines in mind.

## Pull Requests

Pull requests are always welcome. Please:

1. Add thorough, passing unit tests for your changes if applicable.
2. Ensure your changes pass `go vet`, `golint`, and `go test`. These checks
   will be run as part of CI.
3. As much as possible, keep git commits and pull requests small. Small commits
   and small PRs are easier to review.
4. Write [informative, well-formatted](https://chris.beams.io/posts/git-commit/)
   git commit messages.
5. Provide details of your change and why you are making it in the PR
   description, referencing issues or other PRs as appropriate.

Remember that any feedback on a PR is intended to improve your contribution or
help the reviewer better understand it, and respond accordingly by either making
changes or replying with questions or clarifications.

### Dependency Management

clusterlint uses Go modules to manage dependencies, but vendors all dependencies
to help ensure reproducible builds and reduce external dependencies in the build
process.

If you need to add or update a dependency, please do the following:

1. Introduce changes that use the new dependency.
2. Add or update the dependency via `go get -m $module`.
3. Clean up `go.mod` with `go mod tidy`.
4. Vendor code with `go mod vendor`.
5. Add the changes to `go.mod`, `go.sum`, and the `vendor` directory in one
   commit.
6. Add all other changes (i.e., code that uses the new dependency) in a separate
   commit. This will make it easier to review the vendoring change separately
   from the code changes.
7. Submit a pull request containing both the vendoring change and the code
   changes that depend on it. It's important to keep these in a single PR so
   that another developer running `go mod tidy` won't undo your changes.

## Issues

We welcome both bug reports and feature requests as issues. The following
guidelines apply to both, though some parts are more relevant to one or the
other. The maintainers will tag issues as enhancements or bugs upon review.

Before filing an issue, please search to see whether your issue has already been
reported. If it has, consider adding a "thumbs-up" reaction to the issue (NOT a
reply, please!) to indicate to the maintainers that others are seeing the same
problem. Feel free to leave a comment if you can provide additional details not
already given in the issue.

Assuming you have a new issue to report, please:

1. Check that your issue is not already resolved in `master`.
2. Provide as many details as you can on:
   1. What you're trying.
   2. The expected result.
   3. The observed result.
3. Include any relevant environmental information (e.g., Kubernetes version
   you're running against).
