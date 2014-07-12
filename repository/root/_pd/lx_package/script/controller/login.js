exports.Public=true;
exports.login=function(c){
  c.RenderNGPage();
};
exports.auth=function(c){

	c.JsonBody.userName
	c.RenderJson({ok1:false,data:c.JsonBody});
}
