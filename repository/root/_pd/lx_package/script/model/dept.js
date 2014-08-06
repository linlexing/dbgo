exports.GetDeptMenuNodes=function(c,grade,level){
	var dept =c.DBModel("lx_dept")[0];
	var db =dept.DBHelper();
	db.Open();
	try{
		dept.FillWhere(
			'grade like {{strcat ph "\'%\'"}} and gradelevel={{ph}}+1 or\
			{{ph}} like {{strcat "grade" "\'%\'"}} and gradelevel<{{ph}}\
			',grade,level,grade,level);
	}
	finally{
		db.Close();
	}
	return dept.Rows();
};
