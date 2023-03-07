# Changelog

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.1] - 2023-03-07

### Fixed

- `resource/metabase_user`: Fix an issue where an ordering mismatch between the API and state for the groups shows a perpetual diff

## [0.6.0] - 2023-02-17

### Added

* **New Data Source:** `metabase_database`

## [0.5.1] - 2023-02-15

### Fixed

* Fixed documentation for `metabase_database`

## [0.5.0] - 2023-02-15

### Added

* **New Resource:** `metabase_database`

### Fixed

* `resource/metabase_user`: Fix read behaviour to re-create when user is manually deactivated
* `resource/metabase_permissions_group`: Fix read behaviour to re-create when group is manually deleted

### Changed

* Upgraded the SDK to 1.x
* Moved schema definitions to their own package

## [0.4.0] - 2022-10-15

### Added

* **New Resource:** `metabase_permissions_group`
* **New Data Source:** `metabase_permissions_group`

## [0.3.0] - 2022-10-11

### Fixed

* `resource/metabase_user`: Use `nil` for user group memberships instead of an empty array to fix issue with creating users with 0.44 

### Changed

* Upgraded the SDK to 0.14

## [0.2.0] - 2022-08-19

### Added

* **New Resource:** `metabase_user`

### Changed

* Upgraded the SDK to 0.11

## [0.1.0] - 2022-08-08

### Added

* **New Data Source:** `metabase_current_user`
* **New Data Source:** `metabase_user`
