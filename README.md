# Simple Golang Worker Pool

Simple worker pool. Supports adding custom handlers for jobs.

## Examples 

You can view examples in `examples` directory.

## Tests

Tests can be found in `workerpool/workerpool_test.go`.

## Knows issues

1. Handlers MUST NOT panic. Support of panic will be added later.