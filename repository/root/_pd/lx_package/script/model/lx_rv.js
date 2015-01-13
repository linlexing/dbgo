exports.getColumnsAndPK=function(c,name){
	var db = c.Project.NewDBHelper();
	var rev ;
	db.Open();
	try{
		var tab = db.GetData("select datasrc,lab,pk from lx_rv where name={{ph}}",name);
		var row = tab.Row(0);
		var columns = db.QColumnsT(row.datasrc,exports.buildTermplateParam(c));
		var lab ={};
		if(row.lab){
			lab = eval("("+row.lab+")");
		}
		rev = _.map(columns,function(value){
			return {name:value,label:lab[value]};
		});
	}finally{
		db.Close();
	}
	return {columns:rev,pk:row.pk.split(",")};
}
exports.buildTermplateParam=function(c){
	return {
		CurrentGrade:c.CurrentGrade,
		UserName:c.UserName(),
		UserDept:c.Session.Get("user.dept")
	};
}
