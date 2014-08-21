var userModel = require("/model/user.js");
exports.show=function(c){
  c.Render(c.ReadyPJArgs());
};
exports.changepwd=function(c){
	var user = new userModel.model(c,c.UserName());
	if(user.Auth(c.JsonBody.oldPwd)){
		user.ChangePwd(c.JsonBody.newPwd);
		c.RenderJson({ok:true});
	}else{
		c.RenderJson({ok:false});
	}
}
