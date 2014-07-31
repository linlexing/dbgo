userModel = require("/model/user.js");
convert = require("/convert.js")
exports.show=function(c){
	eles = userModel.GetUserElement(c,c.Session.Get("_user.name"));
	fileName = "sys/curele.csv";
	csv = eles.AsCsv();
	if(!c.UserFile.FileExists(fileName) ||
		convert.Bytes2Str(c.UserFile.ReadFile(fileName)) != csv){
		c.UserFile.WriteFile(fileName,convert.Str2Bytes(csv))
	}

	c.RenderNGPage({CurrentUserElement:eles});
}
