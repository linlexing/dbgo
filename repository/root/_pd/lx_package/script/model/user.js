var sha256 = require("/sha256.js");
var convert =require("/convert.js");
exports.model=function(c,userName){
	var m = c.DBModel("lx_user","lx_userrole");
	var db= m.lx_user.DBHelper();
	db.Open();
	try{
		m.lx_user.FillByID(userName);
		if( m.lx_user.RowCount()==0){
			this._exists = false;
		}else{
			this._exists = true;
			row = m.lx_user.Row(0);
			this.Name = row.name;
			this.Password = row.pwd;
			this.Salt= row.salt;
			this.Email = row.email;
			this.DeptName = row.deptname;
			m.lx_userrole.FillWhere("username={{ph}}",userName);
			this.Roles = [];
			if(m.lx_userrole.RowCount() > 0 ){
				this.Roles.push(m.lx_userrole.Row(0).rolename)
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
	var ele =c.DBModel("lx_element").lx_element;
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
	return ele.Rows();
};
