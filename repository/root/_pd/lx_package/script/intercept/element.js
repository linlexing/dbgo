var fmt= require("/fmt.js");
var cnst = require("/const.js");
function clearText(str){
	if(str){
		return str.replace(/[0-9]+\./g,"");
	}else{
		return null;
	}
}
function clearCategory(cate){
	var paths = cate.split("/");
	for(var i in paths){
		paths[i] = clearText(paths[i]);
	}
	return paths.join("/");
}
exports.When =cnst.BFORE;
exports.Intercept = function(c,filter){
	var ele = c.QueryValues()._ele;
	if(ele && ele.length>0){
		ele = ele[0];
		var eleTable=c.DBModel("lx_element")[0];
		var db =eleTable.DBHelper();
		db.Open();
		try{
			eleTable.FillByID(ele);
		}
		finally{
			db.Close();
		}
		if(eleTable.RowCount() >0){
			c.Element = eleTable.Row(0);
			c.Element.label_en_clear = clearText(c.Element.label_en);
			c.Element.label_cn_clear = clearText(c.Element.label_cn);
			c.Element.category_en_clear = clearCategory(c.Element.category_en);
			c.Element.category_cn_clear = clearCategory(c.Element.category_cn);

			c.Element.label_en_clear = c.Element.label_en_clear||c.Element.label_cn_clear;
			c.Element.label_cn_clear = c.Element.label_cn_clear||c.Element.label_en_clear;
			c.Element.category_en_clear = c.Element.category_en_clear||c.Element.category_cn_clear;
			c.Element.category_cn_clear = c.Element.category_cn_clear||c.Element.category_en_clear;
		}
	}
	if(!c.HasResult() && filter.length>0){
		filter[0](c,filter.slice(1));
	}

}
