var sha256 = require("/sha256.js");
var convert =require("/convert.js");
var rand = require("/crypto_rand.js");
var url = require("/url.js");
exports.model=function(c,userName){
	var m = c.DBModel("lx_user","lx_userrole");
	var db= m[0].DBHelper();
	var lx_user = m[0];
	var lx_userrole = m[1];
	db.Open();
	try{
		lx_user.FillByID(userName);
		if( m[0].RowCount()==0){
			this._exists = false;
		}else{
			this._exists = true;
			row = lx_user.Row(0);
			this.Name = row.name;
			this.Password = row.pwd;
			this.Salt= row.salt;
			this.Email = row.email;
			this.DeptName = row.deptname;
			lx_userrole.FillWhere("username={{ph}}",userName);
			this.Roles = [];
			if(lx_userrole.RowCount() > 0 ){
				this.Roles.push(lx_userrole.Row(0).rolename)
			}
		}
	}
	finally{
		db.Close();
	}
	this.Exists = function(){
		return this._exists;
	}
	this.Auth=function(pwd){
		return _.isEqual(this.Password,	sha256.Sum256(this.Salt.concat(convert.Str2Bytes(pwd))));
	}
	this.ChangePwd=function(newPwd){
		var newSalt = convert.NewBytes(10);
		rand.Read(newSalt);
		var newPwd = sha256.Sum256(newSalt.concat(convert.Str2Bytes(newPwd)));
		db.Open();
		try{
			db.Exec("update lx_user set pwd = {{ph}},salt={{ph}} where name={{ph}}",newPwd,newSalt,c.UserName())
		}finally{
			db.Close();
		}
	}
};
exports.BuildUserElementJSFile=function(c){
	var eles =c.DBModel("lx_element")[0];
	var db =eles.DBHelper();
	db.Open();
	try{
		eles.FillWhere(
			"exists(select 1 \
				from lx_userrole a inner join \
					lx_roleele b on a.rolename=b.rolename \
				where a.username={{ph}} and \
					b.elename=dest.name \
			)",c.UserName());
	}
	finally{
		db.Close();
	}
	for(var i =0;i < eles.RowCount();i++){
		row = eles.Row(i);
		if(row.url){
			row.url = url.SetQuery(c.AuthUrl(row.url),{_ele:row.name});
			eles.UpdateRow(i,row);
		}
	}
	var fileName = "sys/curele.js";
	var jsonp = eles.AsJSONP("CurrentUserElement",eles.ColumnNames());
	if(!c.UserFile.FileExists(fileName) ||
		c.UserFile.ReadFileStr(fileName) != jsonp){
		c.UserFile.WriteFileStr(fileName,jsonp);
	}
	return;
};
exports.GetUserDept=function(c,userName){
	dept = c.DBModel("lx_dept")[0];
	db = dept.DBHelper();
	rev = null;
	db.Open();
	try{
		dept.FillWhere("name = (select deptname from lx_user where lx_user.name={{ph}})",userName);
		if(dept.RowCount()>0){
			rev = dept.Row(0);
		}
	}
	finally{
		db.Close();
	}
	return rev
}
