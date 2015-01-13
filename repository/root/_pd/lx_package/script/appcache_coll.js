function collection(c,collName,batch){
	this.c = c;
	this.collName = collName;
	this.batch = batch;
}
collection.prototype.get=function(key){
	return this.c.Project.AppCache.Get(this.collName+"|"+key);
}
collection.prototype.set=function(key,value){
	this.batch.push({opt:"set",key:this.collName+"|"+key,value:value});
}
collection.prototype.delete=function(key){
	this.batch.push({opt:"delete",key:this.collName+"|"+key});
}
collection.prototype.keys=function(keyPrex){
	var prex = this.collName+"|"+keyPrex;
	var iter= this.c.Project.AppCache.PrexIterator(prex);
	var rev = [];
	try{
		while(iter.Next()){
			rev.push(iter.Key().substr(this.collName.length+1));
		}
	}finally{
		iter.Release();
	}
	return rev;
}
exports.collection = collection;
