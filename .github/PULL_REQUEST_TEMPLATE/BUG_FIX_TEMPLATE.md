## Bug Description

<!-- Provide a clear and concise description of the bug -->

## Related Issue

<!-- Link to the bug report issue -->

Fixes #

## Root Cause

<!-- Explain what caused the bug -->

## Changes Made

<!-- List the specific changes made to fix the bug -->

-
-
-

## Type of Bug Fix

- [ ] Traffic leakage / kill-switch failure
- [ ] BPF program loading/attachment error
- [ ] Cgroup management issue
- [ ] Routing table misconfiguration
- [ ] Interface monitoring failure
- [ ] Signal handling issue
- [ ] Memory/resource leak
- [ ] Crash/panic
- [ ] Documentation error
- [ ] Other: <!-- specify -->

## Test Environment

- **OS**: <!-- e.g., Ubuntu 24.04 -->
- **Kernel**: <!-- e.g., 6.8.0-generic -->
- **Go version**: <!-- e.g., 1.24.0 -->
- **clang version**: <!-- e.g., 18.0.0 -->

## Reproduction Steps (Before Fix)

<!-- Steps to reproduce the bug before the fix -->

1.
2.
3.

## Verification Steps (After Fix)

<!-- Steps to verify the bug is fixed -->

1.
2.
3.

## Testing

### Automated Tests

- [ ] Added regression test for this bug
- [ ] All existing tests pass
- [ ] Tested with `go vet` and `staticcheck`

### Manual Testing

<!-- Describe manual testing performed -->

```
<!-- Test results -->
```

## Regression Risk

<!-- Assess the risk of this fix introducing new issues -->

- [ ] Low - Isolated change, well-tested
- [ ] Medium - Touches shared code, needs careful review
- [ ] High - Significant refactoring, extensive testing needed

### Risk Mitigation

<!-- How have you minimized the risk? -->

## Performance Impact

- [ ] No performance impact
- [ ] Performance improvement
- [ ] Slight performance trade-off (justified)

## Breaking Changes

- [ ] This fix introduces breaking changes
- [ ] No breaking changes

## Edge Cases Considered

<!-- List edge cases you've tested -->

- [ ] Interface already down at startup
- [ ] Multiple concurrent netleak sessions
- [ ] Process exits before cleanup
- [ ] Cgroup already exists
- [ ] BPF map already pinned
- [ ] Other: <!-- specify -->

## Checklist

- [ ] Bug is completely fixed (no partial fixes)
- [ ] Fix addresses root cause, not just symptoms
- [ ] Code follows project standards
- [ ] Regression test added
- [ ] Documentation updated (if needed)
- [ ] All CI checks pass
- [ ] Tested on affected kernel versions

## Reviewer Notes

<!-- Specific areas for reviewers to focus on -->
