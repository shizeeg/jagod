#/bin/sh

COMMIT=$(git log -1 --abbrev-commit | grep 'commit' | cut -d ' ' -f2)
COMMITDATE=$(git log -1 --date=short | grep '[Dd]ate'| sed 's|[^0-9-]||g')
CNT=$(git log | grep 'commit' | wc -l | tr -d ' ')
VERSION=$(echo "0.0.${CNT} (GIT ${COMMIT} (${COMMITDATE}))")
sed "s|BOTVERSION = ".*"|BOTVERSION = \"$VERSION\"|g" uname.go > tmp.go
mv tmp.go uname.go
echo "$COMMIT"
echo "$COMMITDATE"
echo "$CNT"
echo "$VERSION"
