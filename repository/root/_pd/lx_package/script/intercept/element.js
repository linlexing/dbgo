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
			var Element = eleTable.Row(0);
			Element.label_en_clear = clearText(Element.label_en);
			Element.label_cn_clear = clearText(Element.label_cn);
			Element.category_en_clear = clearCategory(Element.category_en);
			Element.category_cn_clear = clearCategory(Element.category_cn);

			Element.label_en_clear = Element.label_en_clear||Element.label_cn_clear;
			Element.label_cn_clear = Element.label_cn_clear||Element.label_en_clear;
			Element.category_en_clear = Element.category_en_clear||Element.category_cn_clear;
			Element.category_cn_clear = Element.category_cn_clear||Element.category_en_clear;
			c.SetTag("Element",Element);
		}else{
			throw fmt.Sprintf("the element %s not exists!",ele);
		}
	}
	if(!c.HasResult() && filter.length>0){
		filter[0](c,filter.slice(1));
	}
}
