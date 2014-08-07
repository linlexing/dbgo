exports.Get=function(c,vname){
	var db = c.Project.NewDBHelper();
	var rev ;
	db.Open();
	try{
		rev = db.QueryOne("select value from lx_option where name={{ph}}",vname);
	}finally{
		db.Close();
	}
	return rev;
}
