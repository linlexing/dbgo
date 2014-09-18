fmt = require("/fmt.js");
trigger = require("/trigger.js");
exports.rv=function(c){
	fmt.Printf("websocket at watch.rv,event:%s,message:%s\n",c.ws.Event,c.ws.Message);
}
