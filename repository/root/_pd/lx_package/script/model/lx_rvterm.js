var lx_rv = require("/model/lx_rv.js");
exports.onRender=function(c,args){
	args.rv_info = lx_rv.getColumnsAndPK(c,args.model[args.mdlopt.mdlname].data[0].elename);
}
