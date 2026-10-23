## Why

目录查询不应仅因构造 Manager 就创建用户配置和缓存目录。

## What Changes

增加可选 SkipInitialization 参数，默认行为保持兼容。需要纯读取的 owner 可显式跳过构造期 mkdir；该选项不是权限沙箱，调用者仍需限制 operation 和网络行为。

## Capabilities

### New Capabilities

- `reader-initialization`: 显式无初始化构造。

### Modified Capabilities

## Impact

fit：公共 SDK 拥有构造选项；Template Registry 拥有读取操作边界。不改状态 schema、ref、credential 或现有默认调用。
