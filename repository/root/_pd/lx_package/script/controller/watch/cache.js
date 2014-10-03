var fmt = require("/fmt.js");
var convert = require("/convert.js");
var sha256 = require("/sha256.js");
function sql_get(c,sqlID){
	return c.Project.AppCache.Get("tw.sql|"+sqlID);
}
function sql_put(batch,sqlID,watchSql){
	batch.push({opt:"put",key:"tw.sql|"+sqlID,value:watchSql});
}
function sql_delete(batch,sqlID){
	batch.push({opt:"delete",key:"tw.sql|"+sqlID});
}
function tabsql_get(c,baseTable,sqlID){
	return c.Project.AppCache.Get("tw.tabsql|"+baseTable+"|"+sqlID);
}
function tabsql_getSql(c,baseTable){
	var prex = "tw.tabsql|"+baseTable+"|";
	var iter= c.Project.AppCache.PrexIterator(prex);
	var rev = [];
	try{
		while(iter.Next()){
			rev.push(iter.Key().substr(prex.length));
		}
	}finally{
		iter.Release();
	}
	return rev;
}

function tabsql_put(batch,baseTable,sqlID){
	batch.push({opt:"put",key:"tw.tabsql|"+baseTable+"|"+sqlID,value:true});
}
function tabsql_delete(batch,baseTable,sqlID){
	batch.push({opt:"delete",key:"tw.tabsql|"+baseTable+"|"+sqlID});
}

function sqlrv_get(c,sqlID,rvUUID){
	return c.Project.AppCache.Get("tw.sqlrv|"+sqlID+"|"+rvUUID);
}
function sqlrv_getRvUUID(c,sqlID){
	var prex = "tw.sqlrv|"+sqlID+"|";
	var iter= c.Project.AppCache.PrexIterator(prex);
	var rev = [];
	try{
		while(iter.Next()){
			rev.push(iter.Key().substr(prex.length));
		}
	}finally{
		iter.Release();
	}
	return rev;
}
function sqlrv_count(c,sqlID){
	return c.Project.AppCache.Count("tw.sqlrv|"+sqlID+"|");
}
function sqlrv_put(batch,sqlID,rvUUID){
	batch.push({opt:"put",key:"tw.sqlrv|"+sqlID+"|"+rvUUID,value:true});
}
function sqlrv_delete(batch,sqlID,rvUUID){
	batch.push({opt:"delete",key:"tw.sqlrv|"+sqlID+"|"+rvUUID});
}

function rvsql_get(c,rvUUID){
	return c.Project.AppCache.Get("tw.rvsql|"+rvUUID);
}
function rvsql_put(batch,rvUUID,sqlID){
	batch.push({opt:"put",key:"tw.rvsql|"+rvUUID,value:sqlID});
}
function rvsql_delete(batch,rvUUID){
	batch.push({opt:"delete",key:"tw.rvsql|"+rvUUID});
}
function deleteCache(c,batch,rvUUID){
	var sqlID = rvsql_get(c,rvUUID);
	if(sqlID){
		rvsql_delete(batch,rvUUID);
		//delete the index
		if(sqlrv_get(c,sqlID,rvUUID)){
			sqlrv_delete(batch,sqlID,rvUUID);
		}
		if(sqlrv_count(c,sqlID)==1){
			var watchSql = sql_get(c,sqlID);
			if(watchSql){
				if(tabsql_get(c,watchSql.table,sqlID)){
					tabsql_delete(batch,watchSql.table,sqlID);
				}
				sql_delete(batch,sqlID);
			}
		}
	}
}
function GetRvWatchUrl(c,rvuuid){
	return fmt.Sprintf("/%s/watch/rv/%s",c.Project.Name,rvuuid);
}
function CheckTable(c,baseTable){
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
}
exports.Put = function(c,watchSql,rvUUID){
	//不采用整个object序列化是因为属性的顺序不能保证，会造成相同对象序列化成字符串后，值不相同
	var sqlID = convert.EncodeBase64(sha256.Sum256(convert.Str2Bytes(JSON.stringify(watchSql.btnUrl)+watchSql.table+watchSql.sql+JSON.stringify(watchSql.params))));
	var bw =[];
	//remove pre cache
	deleteCache(c,bw,rvUUID);
	//add web socket map sqlid
	sqlrv_put(bw,sqlID,rvUUID);
	rvsql_put(bw,rvUUID,sqlID);
	//add watch sql list
	tabsql_put(bw,watchSql.table,sqlID);
	sql_put(bw,sqlID,watchSql);
	c.Project.AppCache.BatchWrite(bw);
}
exports.Delete = function(c,rvUUID){
	var bw = [];
	deleteCache(c,bw,rvUUID);
	c.Project.AppCache.BatchWrite(bw);
}
exports.GetWatchSql = function(c,sqlID){
	return sql_get(c,sqlID);
}
exports.GetTableSql = function(c,baseTable){
	CheckTable(c,baseTable);
	return tabsql_getSql(c,baseTable);
}
exports.GetRvUUID = function(c,sqlID){
	return sqlrv_getRvUUID(c,sqlID);
}
exports.GetRvWatchUrl=GetRvWatchUrl;
