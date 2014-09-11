var fmt = require("/fmt.js");
var url = require("/url.js");
function getTableDefine(tab){
	rev = {};
	rev.columns = [];
	var tabColumns = tab.Columns();
	for(var colName in tabColumns){
		rev.columns.push(_.extend(tabColumns[colName],{Name:colName}));
	}
	rev.pk = tab.PK();
	return rev;
}
//从数据库取出客户端所需的数据和结构
function DB2Client(c,mdlname,pk){
	var mdlJS = safeRequire("/model/" + mdlname + ".js");
	var tables={};
	if(mdlJS&&mdlJS.onDB2Client){
		var ev = {modelName:mdlname,pk:pk,data:{}};
		mdlJS.onDB2Client(ev);
		tables = ev.data;
	}else{
		var tab= c.DBModel(mdlname)[0];
		var db =tab.DBHelper();
		db.Open();
		try{
			tab.FillByID(pk);
			tables[mdlname] = tab;
		}finally{
			db.Close();
		}
	}
	var rev = {};
	for(var tabName in tables){
		rev[tabName] = {
			data:tables[tabName].Rows(),
			define:getTableDefine(tables[tabName])
		};
	}
	return rev;
}
//将客户端数据保存到数据库中
function Client2DB(c,mdlname,oldpk,operate,data){
	var mdlJS = safeRequire("/model/" + mdlname + ".js");
	if(mdlJS&&mdlJS.onClient2DB){
		var ev = {modelName:mdlname,oldpk:oldpk,operate:operate,data:data};
		mdlJS.onClient2DB(ev);
	}else{
		var tab = c.DBModel(mdlname)[0];
		var db = tab.DBHelper();
		db.Open();
		try{
			switch(operate){
				case "add":{
					tab.AddRow(data.data[mdlname][0]);
					if(tab.RowCount()==0){
						throw fmt.Sprintf("the pk:%#v 's record not found!",oldpk);
					}
					break;
				}
				case "edit":{
					tab.AddRow(data.originData[mdlname][0]);
					tab.AcceptChange();
					tab.UpdateRow(0,data.data[mdlname][0]);
					break;
				}
				case "delete":{
					tab.AddRow(data.originData[mdlname][0]);
					tab.AcceptChange();
					tab.DeleteRow(0);
					break;
				}
			}
			if(!tab.HasChange()){
				throw fmt.Sprintf("the record same to db,so not change");
			}
			var rcount = tab.Save() ;
			if(rcount != 1){
				throw fmt.Sprintf("the pk:%v 's record can't %s,maybe other user changed the record",oldpk,operate);
			}
		}finally{
			db.Close();
		}
	}

}
function RenderMdlOpt(c,mdlname,operate,fieldsets,pk,args){
	args = args ||{};
	args.mdlopt = {mdlname:mdlname,operate:operate,fieldsets:fieldsets,pk:pk};
	args = _.extend(c.ReadyPJArgs(),args);
	switch(operate){
		case "add":{
			throw fmt.Sprintf("not implement");
			break;
		}
		case "edit":
		case "browse":
		case "delete":{
			if(!pk || pk.length == 0){
				throw fmt.Sprintf("the operate %q need pk value!",operate);
			}
			args.model=DB2Client(c,mdlname,pk);
			if(args.model[mdlname].data.length==0){
				throw fmt.Sprintf("not found model %q record,the pk is :%v",mdlname,pk);
			}
			//fill fieldsets
			for(var fieldName in fieldsets){
				for(var colIndex in args.model[mdlname].define.columns){
					if( args.model[mdlname].define.columns[colIndex].Name == fieldName){
						args.model[mdlname].define.columns[colIndex].Perm = fieldsets[fieldName].perm;
						args.model[mdlname].define.columns[colIndex].Fill = fieldsets[fieldName].fill;
						break;
					}
				}
			}
			args.SaveUrl = c.AuthUrl(url.SetQuery("mdlopt/save",{_n:c.GetTag("Element").name,_pk:pk}));
			break;
		}
		default:{
			throw fmt.Sprintf("the operate %q invalid!",operate);
		}
	}
	var tName = "model/" + mdlname+".html";
	if(c.TemplateExists(tName)){
		c.RenderTemplate(tName,args);
	}else{
		c.RenderTemplate("mdlcommon.html",args);
	}

}

exports.show=function(c){
	var mdlopt = c.DBModel("lx_mdlopt")[0];
	var db = mdlopt.DBHelper();
	var pk = c.QueryValues()._pk;
	db.Open();
	try{
		mdlopt.FillByID(c.GetTag("Element").name);
		if(mdlopt.RowCount() != 1){
			c.RenderError(fmt.Sprintf("the model operate:%q not found!",c.GetTag("Element").name));
		}else{
			var row = mdlopt.Row(0);
			RenderMdlOpt(c,row.mdlname,row.operate,row.fieldsets?eval("("+row.fieldsets+")"):{},pk);
		}
	}finally{
		db.Close();
	}
}
exports.save=function(c){
	var mdlopt = c.DBModel("lx_mdlopt")[0];
	var db = mdlopt.DBHelper();
	var pk = c.QueryValues()._pk;
	var eleName = c.QueryValues()._n;
	db.Open();
	try{
		mdlopt.FillByID(eleName);
	}finally{
		db.Close();
	}
	if(mdlopt.RowCount() != 1){
		c.RenderJson({ok:false,error:fmt.Sprintf("the model operate:%q not found!",eleName)});
	}else{
		var row = mdlopt.Row(0);
		switch(row.operate){
			case "add":
			case "edit":
			case "delete":{
				Client2DB(c,row.mdlname,pk,row.operate,c.JsonBody);
				break;
			}
			default:{
				c.RenderError(fmt.Sprintf("the operate:%q invalid!",row.operate));
				break;
			}
		}
		c.RenderJson({ok:true});
	}
}
