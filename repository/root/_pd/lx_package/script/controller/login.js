var sha256 = require("/sha256.js");
var convert =require("/convert.js");
exports.Public=true;
exports.show=function(c){
  c.RenderNGPage();
};
exports.auth=function(c){
	var lx_user = c.Model("lx_user");
	var ok =false;
	lx_user.FillByID(c.JsonBody.userName);
	if( lx_user.RowCount()==0 || !_.isEqual(lx_user.Row(0).pwd,
		sha256.Sum256(lx_user.Row(0).salt.concat(
			convert.Str2Bytes(c.JsonBody.password)
		)))
	){
		c.RenderJson({ok:false});
	}else{
		c.AuthUrl("home.show");
		c.Session.Set("_username",c.JsonBody.userName);
		c.Session.Set("_deptcode",lx_user.Row(0).deptcode);
		c.RenderJson({ok:true});
	}
}
