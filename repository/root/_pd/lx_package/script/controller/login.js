var userModel =require("/model/user.js");
var grade = require("/grade.js");
exports.Public=true;
exports.show=function(c){
  c.RenderNGPage();
};

exports.auth=function(c){
	var user = new userModel.model(c,c.JsonBody.userName);
	if(user.Exists() &&user.Auth(c.JsonBody.password)){
		dept = c.DBModel("lx_dept").lx_dept;
		db = dept.DBHelper();
		db.Open();
		try{
			var deptLabel_en = "";
			var deptLabel_zh_cn="";
			var deptGrade =grade.GRADE_TAG;
			dept.FillByID(user.DeptName);
			if(dept.RowCount()>0){
				deptLabel_en = dept.Row(0).label_en;
				deptLabel_zh_cn = dept.Row(0).label_zh_cn;
				deptGrade = dept.Row(0).grade;
			}
			c.AuthUrl("home.show");
			c.Session.Set("_user.name",c.JsonBody.userName);
			c.Session.Set("_dept.name",user.DeptName);
			c.Session.Set("_dept.label_en",deptLabel_en);
			c.Session.Set("_dept.label_zh_cn",deptLabel_zh_cn);
			c.Session.Set("_dept.grade",deptGrade);
			c.RenderJson({ok:true});
		}
		finally{
			db.Close();
		}

	}else{
		c.RenderJson({ok:false});
	}
}
