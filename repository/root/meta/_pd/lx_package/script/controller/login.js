exports.Public=true;
exports.login=function(c){
  c.RenderNGPage();
};
exports.auth=function(c){
	var lx_user = c.Model("lx_user");
	var ok =false;
	lx_user.FillByID(c.JsonBody.userName);
	if( lx_user.RowCount() >0 && lx_user.Row(0).){

	}
	console.log('test0');
	var array = new Uint8Array(100);
	if( !array){
		console.log("is null");
	}else{
		console.log("is not null");
	}
	console.log('test1');

	array[0] = 1;
	console.log(array.length);
	c.RenderJson({ok1:false,data:c.JsonBody});
}
