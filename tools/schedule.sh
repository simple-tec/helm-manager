#!/bin/bash
if [ $# != 1 ] ; then 
	echo "USAGE: $0 <prefix>" 
	exit 1; 
fi

nowtime=`date +%Y%m%d%H%M%S`
backupprefix=$1
backupname=${backupprefix}-${nowtime}
echo ${backupname}
token="BeareyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJjb3VudHJ5IjoiQ04iLCJjcmVhdGlvblRpbWUiOjE1ODY0MTE0MDIsImV4cCI6MTU5NTQ5MTAyOSwiZmlyc3ROYW1lIjoic3NvIiwiaWF0IjoxNTk1NDg3NDI5LCJpZCI6ImY0OTI1OTVhLTdhMjUtMTFlYS1hZjhlLWQyOWQ2Nzk4YjAwZiIsImlzcyI6Indpc2UtcGFhcyIsImxhc3RNb2RpZmllZFRpbWUiOjE1OTU0ODY2NjksImxhc3ROYW1lIjoiYWRtaW4iLCJyZWZyZXNoVG9rZW4iOiJiYTQwZjliNS1jY2IxLTExZWEtYTYwMy1lYWUyOGI4ZWI4YmUiLCJzdGF0dXMiOiJBY3RpdmUiLCJ1c2VybmFtZSI6InNzb3Rlc3Ryb290QGVtYWlsLmNvbSJ9.9t37uUpyT4Ss8g7BFnHj4FQ5kSm3n66yv8mlzGO_HVhDiUMiHgkIE1hM5j6yvCGETIn638YEEYMiZqwMU5TtEw"
curl -XPOST http://api-backup-ensaas.bm.wise-paas.com.cn/v1/backup/backup/datacenter/bm/cluster/eks001 -H 'Content-Type: application/json' -H 'Admin: true' -H "Authorization: $token" -d"
{
	\"backupname\": \"${backupname}\",
    \"ismanaged\" : true,
	\"includens\": \"ensaas-service\"
}"

