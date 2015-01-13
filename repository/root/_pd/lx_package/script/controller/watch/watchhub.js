var fmt = require("/fmt.js");
var convert = require("/convert.js");
var sha256 = require("/sha256.js");
var appcache_coll = require("/appcache_coll.js");
function getUniqueID(obj){
	return convert.EncodeBase64(sha256.Sum256(convert.Str2Bytes(
		JSON.stringify(
			_.sortBy(
				_.pairs(obj),
				function(val){
					return val[0];
				}
			)
		)
	)));
}
exports.Unregister = function(c,wsID,watch){
	var watchID = getUniqueID(watch);
	var batch = [];
	var wh_ws2watch = new appcache_coll.collection(c,"wh_ws2watch",batch);
	var wh_watch2ws = new appcache_coll.collection(c,"wh_watch2ws",batch);
	var wh_watch = new appcache_coll.collection(c,"wh_watch",batch);
	var wh_table = new appcache_coll.collection(c,"wh_table",batch);
	wh_ws2watch.delete(wsID+"|"+watchID);
	wh_watch2ws.delete(watchID+"|"+wsID);
	var wsKeys = wh_ws2watch.keys(wsID+"|");
	if(wsKeys.length==1 && wsKeys[0] == wsID+"|"+watchID){
		wh_watch.delete(watchID);
		wh_table.delete(watch.table+"|"+watchID);
	}
	c.Project.AppCache.BatchWrite(batch);
}
exports.Register = function(c,wsID,watch){
	var watchID = getUniqueID(watch);
	var batch = [];
	var wh_ws2watch = new appcache_coll.collection(c,"wh_ws2watch",batch);
	var wh_watch2ws = new appcache_coll.collection(c,"wh_watch2ws",batch);
	var wh_watch = new appcache_coll.collection(c,"wh_watch",batch);
	var wh_table = new appcache_coll.collection(c,"wh_table",batch);
	wh_watch.set(watchID,watch);
	wh_ws2watch.set(wsID+"|"+watchID,true);
	wh_watch2ws.set(watchID+"|"+wsID,true);
	wh_table.set(watch.table+"|"+watchID);
	c.Project.AppCache.BatchWrite(batch);
	return watchID;
}
exports.GetWatch=function(c,watchID){
	var wh_watch = new appcache_coll.collection(c,"wh_watch",null);
	return wh_watch.get(watchID);
}
exports.GetTable=function(c,tableName){
	var wh_table = new appcache_coll.collection(c,"wh_table",null);
	return _.map(wh_table.keys(tableName+"|"),function(val){
		return val.substr(tableName.length+1);
	});
}
exports.GetWSID=function(c,watchID){
	var wh_watch2ws = new appcache_coll.collection(c,"wh_watch2ws",null);
	return _.map(wh_watch2ws.keys(watchID+"|"),function(val){
		return val.substr(watchID.length+1);
	});
}
function GetWSUrl(c,wsID){
	return fmt.Sprintf("/%s/watch/rv/%s",c.Project.Name,wsID);
}
/*function CheckTable(c,baseTable){
	var rev = tabsql_getSql(c,baseTable);
	for(var i in rev){
		var rvs = sqlrv_getRvUUID(c,rev[i]);
		for(var j in rvs){
			var rvUrl = GetRvWatchUrl(c,rvs[j]);
			//删除已经不存在websocket的监视
			if(!c.ChannelExists(rvUrl)){
				var bw =[];
				deleteCache(c,bw,rvs[j]);
				c.Project.AppCache.BatchWrite(bw);
			}
		}
	}
}*/
