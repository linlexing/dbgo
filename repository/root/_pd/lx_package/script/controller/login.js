var userModel =require("/model/user.js");
var grade = require("/grade.js");
var fmt =require("/fmt.js")
exports.Public=true;
exports.show=function(c){
  c.Render();
};
exports.logout=function(c){
	c.Session.Abandon();
	c.RenderRedirection(c.Url("login.show"));
}
exports.auth=function(c){
	var user = new userModel.model(c,c.JsonBody.userName);
	if(user.Exists() &&user.Auth(c.JsonBody.password)){
		userDept =userModel.GetUserDept(c,c.JsonBody.userName);
		if(userDept){
			c.UserName(c.JsonBody.userName);
			c.Session.Set("user.dept" ,userDept);
			userModel.BuildUserElementJSFile(c);
			c.RenderJson({ok:true});
		}else{
			c.RenderJson({ok:false});
		}
	}else{
		c.RenderJson({ok:false});
	}
}
