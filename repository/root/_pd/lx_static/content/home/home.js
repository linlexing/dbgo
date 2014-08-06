var elementTrees;
function clearText(str){
	if(str){
		return str.replace(/[0-9]+\./g,"");
	}else{
		return null;
	}
}
function getPathName(ele){
	rev ={};
	if(ele.category_en != null &&ele.category_en!= ""){
		rev.en = ele.category_en.split("/");
	}
	if(ele.category_cn != null && ele.category_cn!= ""){
		rev.cn = ele.category_cn.split("/");
	}
	return rev;
}
function getLabelName(ele){
	rev ={};
	if(ele.label_en != null &&ele.label_en!= ""){
		rev.en = ele.label_en;
	}
	if(ele.label_cn != null && ele.label_cn!= ""){
		rev.cn = ele.label_cn;
	}
	return rev;
}
function equName(src,dest){
	for(lag in src){
		if(dest[lag] && src[lag] == dest[lag]){
			return true;
		}
	}
	return false;
}
function lgName(src,dest){
	for(lag in src){
		if(dest[lag] && src[lag] > dest[lag]){
			return true;
		}
	}
	return false;
}
//返回第一个语言的内容，en优先
function firstLag(value){
	if(value.en){
		return value.en;
	}else{
		for(var lag in value){
			return value[lag];
			break;
		}
	}
}
function node(parent,pathName,data){
	remove this!
	this.prt = parent;
	this.uid=_.uniqueId("n");
	this.pathName=pathName;
	this.data = data;
	this.children=[];
}
node.prototype.isPathNode=function(){
	return !this.data ;
}

node.prototype.id=function(){
	if(this.isPathNode()){
		return clearText(firstLag(this.pathName));
	}else{
		return this.data.name;
	}
}

node.prototype.displayLabel=function(){
	if(this.data == null){
		return this.pathName;
	}else{
		return getLabelName(this.data);
	}
}
node.prototype.makePath=function(pathName){
	if(_.isEmpty(pathName)){
		return this;
	}
	var foundIdx = 0;
	var rev = null;
	var firstName = {};
	var subName = {};
	for(lag in pathName){
		firstName[lag] = pathName[lag][0];
		if( pathName[lag].length >1){
			subName[lag] = pathName[lag].slice(1);
		}
	}
	for(var i in this.children){
		if(equName(this.children[i].pathName,firstName)){
			if(!_.isEmpty(subName)){
				rev = this.children[i].makePath(subName);
			}else{
				rev = this.children[i];
			}
			break;
		//插入排序
		}else if(lgName(this.children[i].pathName,firstName)){
			foundIdx = i;
			break;
		}
	}
	if(!rev){
		this.children.splice(foundIdx,0,new node(this,firstName,null));
		if(!_.isEmpty(subName)){
			rev = this.children[foundIdx].makePath(subName);
		}else{
			rev = this.children[foundIdx];
		}
	}
	return rev;
}
node.prototype.add=function(data){
	var foundIdx = 0;
	for(var i in this.children){
		if( this.children[i].data && this.children[i].data.name == data.name){
			throw "dup name " + data.name;
		}else if(lgName(this.children[i].displayLabel(),getLabelName(data))){
			foundIdx = i;
			break;
		}
	}
	this.children.splice(foundIdx,0,new node(this,null,data));
}
node.prototype.findById = function(ids){
	if( ids.length == 0) {
		return this;
	}
	var firstName = ids[0];
	var subName = ids.slice(1);
	for(var i in this.children){
		if(this.children[i].id() == firstName){
			return this.children[i].findById(subName);
		}
	}
	throw "not find " + ids.join("/");
}
node.prototype.each= function(cb){
	var v = cb(this);
	if(v){
		return v;
	}
	for(var i in this.children){
		var v = this.children[i].each(cb);
		if(v){
			return v;
		}
	}
}
function toTree(eles){
	var rootNode = new node(null,{en:"root"},null);
	for(var i in eles){
		a = rootNode.makePath(getPathName(eles[i]));
		a.add(eles[i]);
	}
	return rootNode
}
function swith_lag(anode,lag){
	if(lag.toLowerCase() == "zh_cn")
		lag = "cn";
	anode.each(function(v){
		var path=[];
		//can't include root
		for(var p = v;p.prt;p=p.prt){
			path.unshift(p.id())
		}
		v.path = path.join("/");

		v.label = clearText( v.displayLabel()[lag.toLowerCase()]);
		if(!v.label || v.label==""){
			v.label = clearText(firstLag(v.displayLabel()));
		}
	});
}
function clearOtherSelect(root,anode){
	root.each(function(v){
		if(v != anode){
			v.selected = false;
		}
	});
}
function toDeptMenus(deptNodes,currentDept){
	//增加下级
	rev = [];
	parentNodes= [];
	for(var i in deptNodes){
		if(_.str.startsWith(deptNodes[i].grade,currentDept.grade)){
			rev.push(deptNodes[i]);
		}else{
			parentNodes.push(deptNodes[i]);
		}
	}
	if(rev.length>0 && parentNodes.length>0){
		rev.push({divider:true});
	}
	rev = _.union(rev,parentNodes);
	return rev;
}
function swithDeptLag(deptNodes,lag){
	if(lag.toLowerCase() == "zh_cn")
		lag = "cn";
	for(i in deptNodes){
		if(!deptNodes[i].divider){
			if(lag == "en"){
				deptNodes[i].label = deptNodes[i].name + "." + deptNodes[i].label_en;
			}else if(lag == "cn"){
				deptNodes[i].label = deptNodes[i].name + "." + deptNodes[i].label_cn;
			}
			if(!deptNodes[i].label){
				deptNodes[i].label = deptNodes[i].name + "." + (deptNodes[i].label_en||deptNodes[i].label_cn);
			}
		}
	}
}
