var sha256 = require("/sha256.js");
var convert =require("/convert.js");
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
};
exports.GetUserElement=function(c,userName){
	var ele =c.DBModel("lx_element")[0];
	var db =ele.DBHelper();
	db.Open();
	try{
		ele.FillWhere(
			"exists(select 1 \
				from lx_userrole a inner join \
					lx_roleele b on a.rolename=b.rolename \
				where a.username={{ph}} and \
					b.elename=dest.name \
			)",userName);
	}
	finally{
		db.Close();
	}
	return ele.DataTable;
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
