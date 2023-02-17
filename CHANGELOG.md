## 0.7.0 (Unreleased)
## 0.6.0 (February 17, 2023)

FEATURES

* **New Data Source:** `metabase_database`

## 0.5.1 (February 15, 2023)

NOTES:

* Fixed documentation for `metabase_database`

## 0.5.0 (February 15, 2023)

FEATURES

* **New Resource:** `metabase_database`

BUG FIXES:

* resource/metabase_user: Fix read behaviour to re-create when user is manually deactivated
* resource/metabase_permissions_group: Fix read behaviour to re-create when group is manually deleted

NOTES:

* Upgraded the SDK to 1.x
* Moved schema definitions to their own package

## 0.4.0 (October 15, 2022)

FEATURES

* **New Resource:** `metabase_permissions_group`
* **New Data Source:** `metabase_permissions_group`

## 0.3.0 (October 11, 2022)

BUG FIXES:

* resource/metabase_user: Use `nil` for user group memberships instead of an empty array to fix issue with creating users with 0.44 

NOTES:

* Upgraded the SDK to 0.14

## 0.2.0 (August 19, 2022)

FEATURES

* **New Resource:** `metabase_user`

NOTES:

* Upgraded the SDK to 0.11

## 0.1.0 (August  8, 2022)

FEATURES:

* **New Data Source:** `metabase_current_user`
* **New Data Source:** `metabase_user`
