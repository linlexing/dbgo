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
function DB2Client(c,mdlname,operate,pk){
	var mdlJS = safeRequire("/model/" + mdlname + ".js");
	var rev={};
	if(mdlJS&&mdlJS.onDB2Client){
		var ev = {modelName:mdlname,operate:operate,pk:pk,data:{}};
		mdlJS.onDB2Client(ev);
		rev = ev.data;
	}else{
		var tab= c.DBModel(mdlname)[0];
		var db =tab.DBHelper();
		var row ;
		db.Open();
		try{
			switch(operate){
				case "add":
					row = tab.NewRow();
					var pkFields = tab.PK();
					for(var pkIndex in pk){
						row[pkFields[pkIndex]] = pk[pkIndex];
					}
					break;
				case "browse":
				case "delete":
				case "edit":
					tab.FillByID(pk);
					row = tab.Row(0);
					break;
			}
			rev[mdlname] = {data:[row],define:tab};
		}finally{
			db.Close();
		}
	}
	for(var tabName in rev){
		rev[tabName].define = getTableDefine(rev[tabName].define);
	}
	return rev;
}
//将客户端数据保存到数据库中
function Client2DB(c,eleName,mdlname,oldpk,operate,data){
	var mdlJS = safeRequire("/model/" + mdlname + ".js");
	var dataSet ;
	var db;
	var mainTable;
	var userDept = c.Session.Get("user.dept");
	var optInfo = {
		UserName:c.UserName(),
		Element:eleName,
		Grade:c.CurrentGrade,
		ClientAddr:c.ClientAddr,
		DeptName:userDept.name,
		DeptLabel_en:userDept.label_en,
		DeptLabel_cn:userDept.label_cn,
		Time:new Date()
	};
	try{
		if(mdlJS&&mdlJS.onClient2DB){
			var ev = {modelName:mdlname,oldpk:oldpk,operate:operate,data:data,dataSet:null};
			mdlJS.onClient2DB(ev);
			dataSet = ev.dataSet;
			mainTable = dataSet[0];
			db = mainTable.DBHelper();
		}else{
			dataSet = c.DBModel(mdlname);
			mainTable = dataSet[0];
			db = mainTable.DBHelper();
			db.Open();
			switch(operate){
				case "add":{
					mainTable.AddRow(data.data[mdlname][0]);
					if(tab.RowCount()==0){
						throw fmt.Sprintf("the pk:%#v 's record not found!",oldpk);
					}
					break;
				}
				case "edit":{
					mainTable.AddRow(data.originData[mdlname][0]);
					mainTable.AcceptChange();
					mainTable.UpdateRow(0,data.data[mdlname][0]);
					break;
				}
				case "delete":{
					mainTable.AddRow(data.originData[mdlname][0]);
					mainTable.AcceptChange();
					mainTable.DeleteRow(0);
					break;
				}
			}
		}
		if( dataSet.length == 1 && !mainTable.HasChange()){
			throw fmt.Sprintf("the record same to db,so not change");
		}
		for(var i in dataSet){
			if(dataSet[i].HasChange()){
				var chgs = dataSet[i].GetChange();
				_.each(chgs.DeleteRows,function(val){
					onRecordChange(optInfo,"delete",val.OriginData,val.Data);
					});
				_.each(chgs.UpdateRows,function(val){
					onRecordChange(optInfo,"update",val.OriginData,val.Data);
					});
				_.each(chgs.InsertRows,function(val){
					onRecordChange(optInfo,"insert",val.OriginData,val.Data);
					});
				dataSet[i].Save();
			}
		}
	}finally{
		db.Close();
	}

}
function onRecordChange(optInfo,operate,originData,newData){
	fmt.Printf("write log:%#v,operate:%s\n%s",optInfo,operate,JSON.stringify(optInfo));
}
function convertFillValue(c,dataType,fillValue){
	if(fillValue===undefined){
		return undefined;
	}
	if(fillValue){
		return eval("("+fillValue+")");
	}else{
		return null;
	}
}
function RenderMdlOpt(c,mdlname,operate,fieldsets,pk,args){
	args = args ||{};
	args.mdlopt = {mdlname:mdlname,operate:operate,fieldsets:fieldsets,pk:pk};
	args = _.extend(c.ReadyPJArgs(),args);
	switch(operate){
		case "add":{
			args.model=DB2Client(c,mdlname,operate,pk);
			if(args.model[mdlname].data.length==0){
				throw fmt.Sprintf("the model %q default new record not found,the pk is :%v",mdlname,pk);
			}
			//fill fieldsets
			for(var fieldName in fieldsets){
				for(var colIndex in args.model[mdlname].define.columns){
					if( args.model[mdlname].define.columns[colIndex].Name == fieldName){
						args.model[mdlname].define.columns[colIndex].Perm = fieldsets[fieldName].perm;
						var v = convertFillValue(c,args.model[mdlname].define.columns[colIndex].DataType,fieldsets[fieldName].fill);
						if(v !== undefined){
							args.model[mdlname].data[0][args.model[mdlname].define.columns[colIndex].Name]= v;
						}
						break;
					}
				}
			}
			args.SaveUrl = c.AuthUrl(url.SetQuery("mdlopt/save",{_n:c.GetTag("Element").name,_pk:pk}));
			break;
		}
		case "edit":
		case "browse":
		case "delete":{
			if(!pk || pk.length == 0){
				throw fmt.Sprintf("the operate %q need pk value!",operate);
			}
			args.model=DB2Client(c,mdlname,operate,pk);
			if(args.model[mdlname].data.length==0){
				throw fmt.Sprintf("not found model %q record,the pk is :%v",mdlname,pk);
			}
			//fill fieldsets
			for(var fieldName in fieldsets){
				for(var colIndex in args.model[mdlname].define.columns){
					if( args.model[mdlname].define.columns[colIndex].Name == fieldName){
						args.model[mdlname].define.columns[colIndex].Perm = fieldsets[fieldName].perm;
						args.model[mdlname].define.columns[colIndex].Fill =convertFillValue(c,fieldsets[fieldName].fill);
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
	var eleName = c.QueryValues("_n");
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
				Client2DB(c,eleName,row.mdlname,pk,row.operate,c.JsonBody);
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
