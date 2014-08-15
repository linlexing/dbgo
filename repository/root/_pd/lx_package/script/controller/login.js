var userModel =require("/model/user.js");
var grade = require("/grade.js");
var fmt =require("/fmt.js")
exports.Public=true;
exports.show=function(c){
  c.Render();
};

exports.auth=function(c){
	var user = new userModel.model(c,c.JsonBody.userName);
	if(user.Exists() &&user.Auth(c.JsonBody.password)){
		userDept =userModel.GetUserDept(c,c.JsonBody.userName);
		if(userDept){
			c.AuthUrl("home.default");
			c.Session.Set("user.name",c.JsonBody.userName);
			c.Session.Set("user.dept" ,userDept);
			c.RenderJson({ok:true});
		}else{
			c.RenderJson({ok:false});
		}
	}else{
		c.RenderJson({ok:false});
	}
}
