userModel = require("/model/user.js");
convert = require("/convert.js");
deptModel = require("/model/dept.js");
grade = require("/grade.js");
opt = require("/model/option.js");
exports.show=function(c){
	eles = userModel.GetUserElement(c,c.UserName);
	for(var i =0;i < eles.RowCount();i++){
		row = eles.Row(i);
		if(row.url){
			c.AuthUrl(row.url);
		}
		eles.UpdateRow(i,row);
	}
	fileName = "sys/curele.js";
	jsonp = eles.AsJSONP("CurrentUserElement",eles.ColumnNames());
	if(!c.UserFile.FileExists(fileName) ||
		c.UserFile.ReadFileStr(fileName) != jsonp){
		c.UserFile.WriteFileStr(fileName,jsonp);
	}
	sdept = c.Session.Get("user.dept");
	var defaultUrl = opt.Get(c,"home_default")||"home/default";
	c.Render({
		deptData:deptModel.GetDeptMenuNodes(c,sdept.grade,sdept.gradelevel),
		defaultUrl:defaultUrl
	});
}
exports.switch_dept=function(c){
	var newDept =c.JsonBody.dept;
	var originGrade = userModel.GetUserDept(c,c.Session.Get("user.name")).grade;
	rev = {};
	if(newDept.grade.indexOf(originGrade)==0){
		rev.ok = true;
		c.Session.Set("user.dept" ,newDept);
		rev.deptData =deptModel.GetDeptMenuNodes(c,newDept.grade,newDept.gradelevel);
	}else{
		rev.ok = false;
		rev.error = "old:" + originGrade + ",new:" + newGrade;
	}
	c.RenderJson(rev);
}
exports.default=function(c){
	c.RenderMDPage({Title_en:"welcome",Title_cn:"欢迎"});
}
