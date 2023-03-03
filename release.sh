#!/bin/bash
 
set -e

current=$(cat version.go | tail -1 | head -1 | awk -F= '{print $2}')
echo "Current version is:"
echo $current
echo -ne "Enter new version: "
read new_version

(
cat <<EOF
package caddy_saml_sso

const version = "$new_version"
EOF
) > version.go

git add version.go
git commit -m 'bump version'
git tag "v$new_version"
git push origin
git push origin v$new_version

make clean release

gh release create v0.0.2 --notes "v$version" ./caddy.*
