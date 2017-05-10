export VERSION=v.0.1.0
echo version=$VERSION
export LDFLAGS=" -w -s -X main._VERSION_=%VERSION%"

echo start install sever ...
go install server/twmain

rm -f config/*.json
cp src/server/config/*.json config/
echo complate

