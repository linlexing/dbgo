var convert = require("/convert.js");
var sha256 = require("/sha256.js");
function sql_get(c,baseTable,sqlID){
	return c.Project.AppCache.Get("tw.sql|"+baseTable+"|"+sqlID);
}
function sql_getTableSql(c,baseTable){
	var prex = "tw.sql|"+baseTable+"|";
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

function sql_put(batch,baseTable,sqlID,watchSql){
	batch.push({opt:"put",key:"tw.sql|"+baseTable+"|"+sqlID,value:watchSql});
}
function sql_delete(batch,baseTable,sqlID){
	batch.push({opt:"delete",key:"tw.sql|"+baseTable+"|"+sqlID});
}

function sqlrv_get(c,sqlID,rvUUID){
	return c.Project.AppCache.Get("tw.sqlrv|"+sqlID+"|"+rvUUID);
}
function sql_getRvUUID(c,sqlID){
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
function sqlrv_has(c,sqlID){
	return c.Project.AppCache.HasPrex("tw.sqlrv|"+sqlID+"|");
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
	var preSqlID = rvsql_get(c,rvUUID);
	if(preSqlID){
		//delete the index
		if(sqlrv_get(c,preSqlID,rvUUID)){
			sqlrv_delete(batch,preSqlID,rvUUID);
		}
		if(!sqlrv_has(c,preSqlID)){
			if(sql_get(c,baseTable,preSqlID)){
				sql_delete(batch,baseTable,preSqlID);
			}
		}
	}
}
exports.Put = function(c,baseTable,watchSql,rvUUID){
	//不采用整个object序列化是因为属性的顺序不能保证，会造成相同对象序列化成字符串后，值不相同
	var sqlID = convert.EncodeBase64(sha256.Sum256(convert.Str2Bytes(watchSql.sql+JSON.stringify(watchSql.params))));
	var bw =[];
	//remove pre cache
	deleteCache(c,bw,rvUUID);
	//add web socket map sqlid
	sqlrv_put(bw,sqlID,rvUUID);
	rvsql_put(bw,rvUUID,sqlID);
	//add watch sql list
	sql_put(bw,baseTable,sqlID,watchSql);
	c.Project.AppCache.BatchWrite(bw);
}
exports.Delete = function(c,rvUUID){
	var bw = [];
	deleteCache(c,bw,rvUUID);
	c.Project.AppCache.BatchWrite(bw);
}
exports.GetWatchSql = function(c,baseTable,sqlID){
	return sql_get(c,baseTable,sqlID);
}
exports.GetTableSql = function(c,baseTable){
	return sql_getTableSql(c,baseTable);
}
exports.GetRvUUID = function(c,sqlID){
	return sql_getRvUUID(c,sqlID);
}
