var fmt = require("/fmt.js");
var cache = require("watch/cache.js");
exports.rv=function(c){
	var rvUUID = c.TagPath;
	switch(c.ws.Event){
		case "close":
			cache.Delete(c,c.TagPath);
			break;
		case "open":
			var watchSql = c.Session.Get("watchSql|"+c.TagPath);
			if(watchSql){
				cache.Put(c,watchSql,c.TagPath);
			}
			break;
	}
	fmt.Printf("websocket at watch.rv,event:%s,message:%s,tagPath:%s\n",c.ws.Event,c.ws.Message,c.TagPath);
}

