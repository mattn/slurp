To contribute to Slurp, please consider the following guidelines.

1. Please always open an issue and discuss it first if you want to make it part of the main repository, this helps pervening duplication of effort, and ensures that the idea fits inside the goals of Slurp project. It also helps refining the deisgn before code is written and results in a simpler code reivew.

2. If your code depends on third-party libraries, write it as an indepnt tool that should go under `github.com/slurp-contrib`.

3. Don't use the `log` package, use the Context (s.C) for logging.

4. When building _stages_, 

  4.1 Deliver or Destroy, that is, if you are not passing a file, Close it.

  4.2 Always assume that the downstream is streaming, so return the file as soon as possible.


5. Use `gofmt -w -s` before creating pull requests.

