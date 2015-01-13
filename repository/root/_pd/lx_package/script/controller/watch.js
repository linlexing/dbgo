var fmt = require("/fmt.js");
var watchhub= require("watch/watchhub.js");
exports.rv=function(c){
	var rvUUID = c.TagPath;
	switch(c.ws.Event){
		case "close":
			var watchList = c.Session.Get("watch|"+c.TagPath);
			if(watchList){
				_.each(watchList,function(val){
					var watchDefine = c.Session.Get(val);
					if(watchDefine){
						fmt.Printf("Unregister %s,table:%s\n",c.TagPath,watchDefine.table);
						watchhub.Unregister(c,c.TagPath,watchDefine);
					}
				});
			}
			break;
		case "open":
			var watchList = c.Session.Get("watch|"+c.TagPath);
			if(watchList){
				_.each(watchList,function(val){
					var watchDefine = c.Session.Get(val);
					if(watchDefine){
						fmt.Printf("Register %s,table:%s\n",c.TagPath,watchDefine.table);
						watchhub.Register(c,c.TagPath,watchDefine);
					}
				});
			}
			break;
	}
}

