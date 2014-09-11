userModel = require("/model/user.js");
convert = require("/convert.js");
deptModel = require("/model/dept.js");
grade = require("/grade.js");
opt = require("/model/option.js");
exports.show=function(c){
	sdept = c.Session.Get("user.dept");
	var defaultUrl = opt.Get(c,"home_default")||"home/default?_ele=home_default";
	c.Render({
		deptData:deptModel.GetDeptMenuNodes(c,sdept.grade,sdept.gradelevel),
		defaultUrl:defaultUrl
	});
}
exports.switch_dept=function(c){
	var newDept =c.JsonBody.dept;
	var originGrade = userModel.GetUserDept(c,c.UserName()).grade;
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
	c.RenderPJMDPage();
}
