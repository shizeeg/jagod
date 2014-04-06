#/bin/sh

COMMIT=$(git log -1 --abbrev-commit | grep 'commit' | sed 's|commit\s*||g')
COMMITDATE=$(git log -1 --date=short | grep '[Dd]ate'| sed 's|[Dd]ate:\s*||g')
CNT=$(git log | grep 'commit' | wc -l)
VERSION=$(echo "0.0.${CNT} (GIT" "${COMMIT} (${COMMITDATE}))")
sed "s|BOTVERSION = ".*"|BOTVERSION = \"$VERSION\"|g" uname.go > tmp.go
mv tmp.go uname.go
