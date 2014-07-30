userModel = require("/model/user.js");
exports.show=function(c){
	eles = userModel.GetUserElement(c,c.Session.Get("_user.name"));
	c.RenderNGPage({CurrentUserElement:eles});
}
