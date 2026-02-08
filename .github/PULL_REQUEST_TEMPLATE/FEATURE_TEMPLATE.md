## Feature Description

<!-- Provide a clear and concise description of the new feature -->

## Related Issue

<!-- Link to the feature request issue -->

Closes #

## Motivation

<!-- Why is this feature needed? What problem does it solve? -->

## Proposed Changes

<!-- Detailed description of the implementation -->

### API / CLI Changes

<!-- If this adds/modifies the command-line interface -->

```
# New usage examples
```

### BPF Changes

<!-- If this modifies the BPF programs or maps -->

```c
/* New/modified BPF structures or programs */
```

### Architecture Changes

<!-- If this changes internal architecture -->

## Implementation Details

<!-- Technical details about the implementation -->

### Core Changes

-
-

### BPF / Kernel Interaction

<!-- Explain any new BPF programs, maps, or kernel interactions -->

## Design Decisions

<!-- Explain key design choices and trade-offs -->

### Alternatives Considered

1. **Alternative 1**:
   - Pros:
   - Cons:

2. **Alternative 2**:
   - Pros:
   - Cons:

### Why This Approach?

<!-- Justify your design choice -->

## Usage Examples

### Basic Usage

```bash
# Example showing how to use the new feature
```

## Test Environment

- **OS**: <!-- e.g., Ubuntu 24.04 -->
- **Kernel**: <!-- e.g., 6.8.0-generic -->
- **Go version**: <!-- e.g., 1.24.0 -->
- **clang version**: <!-- e.g., 18.0.0 -->

## Testing

### Test Coverage

- [ ] Unit tests for new Go code
- [ ] BPF program tested (if applicable)
- [ ] Edge case tests
- [ ] Error handling tests
- [ ] Integration tests with real interface (if applicable)

### Test Results

```
<!-- Test output -->
```

## Performance

### Benchmarks

<!-- Required for performance-related features -->

```
<!-- Benchmark results -->
```

### Performance Characteristics

- **Latency impact**:
- **Memory impact**:
- **BPF instruction count** (if applicable):

## Documentation

- [ ] Code comments for public functions
- [ ] README.md updated
- [ ] Usage examples provided
- [ ] Man page updated (if applicable)

## Backward Compatibility

- [ ] Fully backward compatible
- [ ] Deprecates old behavior (migration guide provided)
- [ ] Breaking change (justified and documented)

## Kernel Compatibility

<!-- What kernel versions does this feature require? -->

- **Minimum kernel version**:
- **New BPF features used** (if any):

## Security Considerations

<!-- Any security implications of this feature? -->

- [ ] No security implications
- [ ] Security review completed
- [ ] Input validation added
- [ ] Privileges documented

## Dependencies

- [ ] No new dependencies
- [ ] New dependencies added (justified below)

## Future Work

<!-- Related features or improvements for future PRs -->

-
-

## Checklist

- [ ] Feature is complete and tested
- [ ] Code follows project standards
- [ ] All tests pass
- [ ] Documentation complete
- [ ] Examples provided
- [ ] Backward compatibility maintained (or justified)
- [ ] CI checks pass
- [ ] Tested on multiple kernel versions

## Reviewer Notes

<!-- Areas you'd like reviewers to focus on -->

### Review Focus Areas

-
-

### Open Questions

-
-
