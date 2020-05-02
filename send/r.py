import time, random, sys, json, codecs, os
j = codecs.open("data.json","r","utf-8")
module = "go get github.com/slayer/autorestart && go get github.com/levigross/grequests && go get github.com/valyala/fasthttp"
squad = json.load(j)
for mid in squad["login"]:
	token = squad["login"][mid]["token"]
	os.system('screen -S '+mid+' -X quit')
	os.system('rm -rf clone/'+mid)
	time.sleep(2)
	os.system('screen -dmS '+mid)
	os.system('screen -r '+mid+' -X stuff "go run sb.go {}\n"'.format(token))