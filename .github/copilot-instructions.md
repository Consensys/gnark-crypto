---
applyTo: "**"
---
# Project general coding standards

## Coding Style
- Follow https://google.github.io/styleguide/go/ for Go code style.

## Templating
- Most files in this project are generated using templates. Make sure to modify the actual template and run `go generate ./...` to update the generated files instead of modifying them directly.
- The templates are located in the `internal/generator` and `field/generator` directories.

## Testing
- Since most files are generated, ensure that tests are also updated in the templates.
- Also ensure to retest all modified (use `git status` to see modified files) generated files to confirm that they work as expected.
- You may use `go test -short` to speed up testing of generated files.
- Write unit tests for all new features and bug fixes.

## Benchmarking

- Use `go test -short -benchmem -count=4 -bench .` to run benchmarks.

## Performance and optimization

- Focus on efficiency: minimize memory allocations and CPU usage, and pick the most efficient algorithms and data structures.
- Profile the code using Go's built-in profiling tools to identify bottlenecks before optimizing.