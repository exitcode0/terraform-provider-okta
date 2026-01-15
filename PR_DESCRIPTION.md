# Fix: Unified resource handling in okta_resource_set

## Summary

This PR fixes persistent drift issues with `okta_resource_set` by using a unified `resources` field that properly handles both URL-based and ORN-based resources, matching the actual behavior of the Okta API.

## Issues Fixed

- Fixes #2224 - ORN resources causing perpetual drift
- Fixes #2006 - Support case resources requiring `ignore_changes` workaround
- Fixes #1991 - ORN path changes always detected

## Problem

The current implementation artificially separates resources into two fields (`resources` and `resources_orn`), but the Okta API returns all resources in a single unified list. This mismatch causes several issues:

1. **Drift with ORN resources**: When users put ORN values in the `resources` field, `flattenResourceSetResourcesLinks()` only checks `if res.Links != nil`, skipping ORN resources (which have `Links == nil`), resulting in an empty set and perpetual drift.

2. **Artificial separation**: The API doesn't distinguish between URL and ORN resources - they're all in the same response. The provider shouldn't either.

3. **Can't mix resource types**: The `ExactlyOneOf` constraint prevents mixing URL and ORN resources, even though the API supports it.

4. **User confusion**: Users must know whether a resource will have `_links` or `orn` populated and choose the correct field.

## Solution

### Key Changes

1. **Unified `resources` field**: Changed from optional with `ExactlyOneOf` constraint to required, accepting both URLs and ORNs.

2. **Smart flattening function**: `flattenResourceSetResources()` checks both `orn` and `_links` fields:
   ```go
   func flattenResourceSetResources(resources []*sdk.ResourceSetResource) *schema.Set {
       var arr []interface{}
       for _, res := range resources {
           var resourceIdentifier string
           
           // Prefer ORN if available (canonical identifier)
           if res.Orn != "" {
               resourceIdentifier = res.Orn
           } else if res.Links != nil {
               // Fall back to link-based identifier
               resourceIdentifier = utils.LinksValue(res.Links, "self", "href")
           }
           
           if resourceIdentifier != "" {
               arr = append(arr, encodeResourceSetResourceLink(resourceIdentifier))
           }
       }
       return schema.NewSet(schema.HashString, arr)
   }
   ```

3. **Simplified schema**: Removed `resources_orn` field entirely.

4. **Updated tests**: 
   - Modified existing ORN test to use unified `resources` field
   - Added `TestAccResourceOktaResourceSet_Issue_2224_orn_no_drift` to verify no drift
   - Added `TestAccResourceOktaResourceSet_mixed_resources` to verify mixing URLs and ORNs

5. **Enhanced documentation**: Updated examples to show ORN usage and mixed resources.

## Benefits

✅ **No drift for ORN resources** - Properly reads back both URL and ORN resources  
✅ **Matches API behavior** - Single unified list like the API returns  
✅ **Can mix resource types** - URLs and ORNs in the same resource set  
✅ **Simpler for users** - One field, no need to know implementation details  
✅ **Support cases work** - No more `ignore_changes` workaround needed  
✅ **Cleaner code** - Removed artificial separation logic  

## Breaking Changes

⚠️ **This is a breaking change** - The `resources_orn` field has been removed.

### Migration Path

Users currently using `resources_orn` should migrate to `resources`:

**Before:**
```hcl
resource "okta_resource_set" "example" {
  label       = "My Resource Set"
  description = "Example"
  resources_orn = [
    "orn:okta:directory:00o123:users",
  ]
}
```

**After:**
```hcl
resource "okta_resource_set" "example" {
  label       = "My Resource Set"
  description = "Example"
  resources = [
    "orn:okta:directory:00o123:users",
  ]
}
```

The migration is straightforward - just rename the field from `resources_orn` to `resources`.

## Testing

### Unit Tests
- ✅ `TestAccResourceOktaResourceSet_crud` - Basic CRUD operations
- ✅ `TestAccResourceOktaResourceSet_Issue1097_Pagination` - Pagination with 201 resources
- ✅ `TestAccResourceOktaResourceSet_Issue_1735_drift_detection` - Drift detection
- ✅ `TestAccResourceOktaResourceSet_Issue_1991_orn_handling` - ORN resource handling (updated)
- ✅ `TestAccResourceOktaResourceSet_Issue_2224_orn_no_drift` - Verify no drift with ORNs (new)
- ✅ `TestAccResourceOktaResourceSet_mixed_resources` - Mixed URL and ORN resources (new)

### Manual Testing Scenarios

1. **ORN resources (Issue #2224)**:
   ```hcl
   resource "okta_resource_set" "test" {
     label       = "Test ORN"
     description = "Testing ORN resources"
     resources = [
       "orn:okta:directory:00o123:users",
     ]
   }
   ```
   - Apply twice, verify no drift

2. **Support cases (Issue #2006)**:
   ```hcl
   resource "okta_resource_set" "support" {
     label       = "Support Cases"
     description = "Support cases"
     resources = [
       "orn:okta:support:00o123:cases",
     ]
   }
   ```
   - Apply twice, verify no drift
   - No `ignore_changes` needed

3. **Mixed resources**:
   ```hcl
   resource "okta_resource_set" "mixed" {
     label       = "Mixed"
     description = "Mixed resources"
     resources = [
       "https://org.okta.com/api/v1/apps",
       "orn:okta:directory:00o123:users",
     ]
   }
   ```
   - Verify both resources are properly managed

## Documentation Updates

- Updated resource description to mention ORN support
- Added example for ORN resources (support cases)
- Added example for mixed URL and ORN resources
- Updated schema documentation to reflect unified field

## Backwards Compatibility

This is a **breaking change** requiring a major version bump. The `resources_orn` field has been removed in favor of the unified `resources` field.

### Deprecation Strategy (Alternative Approach)

If we want to maintain backwards compatibility for one more release, we could:

1. Keep both fields temporarily
2. Mark `resources_orn` as deprecated
3. Make both fields optional (not `ExactlyOneOf`)
4. In the read function, populate whichever field the user configured
5. Remove `resources_orn` in the next major version

However, I recommend the clean break approach in this PR because:
- The migration is trivial (just rename the field)
- Keeping both fields maintains the confusing separation
- Users on `resources_orn` are already experiencing issues

## Related Issues

- #2224 - ORN resources don't return `_links` attribute
- #2006 - Add support cases to `okta_resource_set`
- #1991 - `okta_resource_set` always detects change in ORN path
- #1504 - Related to ORN handling

## Checklist

- [x] Code changes implement unified resource handling
- [x] Tests updated to use unified `resources` field
- [x] New tests added for drift detection and mixed resources
- [x] Documentation updated with examples
- [x] Schema documentation reflects breaking change
- [ ] CHANGELOG.md updated (pending PR number)
- [ ] Migration guide prepared for users

## Additional Notes

This fix addresses the root cause rather than treating symptoms. The previous `resources_orn` approach was a workaround that didn't match the API's actual behavior. This unified approach is simpler, more correct, and eliminates the drift issues entirely.
