var fmt = require("/fmt.js");
var url = require("/url.js");
var watchhub = require("watch/watchhub.js");
function fireEvent(c,mdlname,evName,evData){
	var mdlJS = safeRequire("/model/" + mdlname + ".js");
	if(mdlJS&&mdlJS[evName]){
		mdlJS[evName](c,evData);
		return evData;
	}else{
		return false;
	}

}
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
	var rev={};
	var ev = {modelName:mdlname,operate:operate,pk:pk,data:{}};
	if(fireEvent(c,mdlname,"onDB2Client",ev)){
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
		var ev = {modelName:mdlname,oldpk:oldpk,operate:operate,data:data,dataSet:null};
		if(fireEvent(c,mdlname,"onClient2DB",ev)){
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
				var deleteRows = chgs.DeleteRows.concat();
				var updateRows = chgs.UpdateRows.concat();
				var insertRows = chgs.InsertRows.concat();
				_.each(deleteRows,function(val){
					onBeforeChange(c,db,optInfo,dataSet[i],"delete",val);
					});
				_.each(updateRows,function(val){
					onBeforeChange(c,db,optInfo,dataSet[i],"update",val);
					});
				_.each(insertRows,function(val){
					onBeforeChange(c,db,optInfo,dataSet[i],"insert",val);
					});
				dataSet[i].Save();
				_.each(deleteRows,function(val){
					onAfterChange(c,db,optInfo,dataSet[i],"delete",val);
					});
				_.each(updateRows,function(val){
					onAfterChange(c,db,optInfo,dataSet[i],"update",val);
					});
				_.each(insertRows,function(val){
					onAfterChange(c,db,optInfo,dataSet[i],"insert",val);
					});
			}
		}
	}finally{
		db.Close();
	}

}
function onBeforeChange(c,db,optInfo,table,operate,rowAgent){
	var baseTable = table.TableName;
	var sqlIDs = watchhub.GetTable(c,baseTable);
	var pkFields = table.PK();
	rowAgent.watch = [];
	var oldPk=[];
	switch(operate){
		case "update":
		case "delete":
			for(var i in pkFields){
				oldPk.push(rowAgent.OriginData[table.ColumnIndex(pkFields[i])]);
			}
			break;
	}
	for(var i in sqlIDs){
		var watch = {
			sqlID:sqlIDs[i],
			watchSql:watchhub.GetWatch(c,sqlIDs[i])
		};
		switch(operate){
			case "update":
			case "delete":
				var oldData = db.GetData(watch.watchSql.sql,oldPk.concat(watch.watchSql.params));
				//print oldpk
				console.log("opt:"+operate+",originData:"+oldData.RowCount());
				if(oldData.RowCount()>0){
					watch.originData = oldData.Row(0);
				}
				break;
		}
		rowAgent.watch.push(watch);
	}
}
function GetWSUrl(c,wsID){
	return fmt.Sprintf("/%s/watch/rv/%s",c.Project.Name,wsID);
}
function sendSqlIDMessage(c,sqlID,message){
	var rvuuids = watchhub.GetWSID(c,sqlID);
	_.each(rvuuids,function(val){
		//由于每个rv均有唯一的url，所以，Broadcast 这里当做send使用
		c.Broadcast(GetWSUrl(c,val),JSON.stringify(message));
	});

}
function buildBtnUrl(c,btnUrls,pkFields,pkValues,table){
	return _.map(btnUrls,function(val){
		return c.AuthUrl(url.SetQuery(val,{
			_pk:_.map(pkFields,function(val,idx){
				return table.EncodeString(val,pkValues[idx]);
			})
		}));
	});
}
function onAfterChange(c,db,optInfo,table,operate,rowAgent){
	var baseTable = table.TableName;
	var pkFields = table.PK();
	var newPk=[];
	switch(operate){
		case "update":
		case "insert":
			for(var i in pkFields){
				newPk.push(rowAgent.Data[table.ColumnIndex(pkFields[i])]);
			}
			break;
	}
	for(var i in rowAgent.watch){
		var watch = rowAgent.watch[i];
		switch(operate){
			case "insert":
			case "update":
				var newData = db.GetData(watch.watchSql.sql,newPk.concat(watch.watchSql.params));
				if(newData.RowCount()>0){
					watch.data = newData.Row(0);
				}
				break;
		}
		//判断最终每个监视视图的增删改操作，并发送websocket消息
		switch(operate){
			case "insert":
				if(watch.data){
					sendSqlIDMessage(c,watch.sqlID,{
						opt:"insert",
						data:watch.data,
						btnUrl:buildBtnUrl(c,watch.watchSql.btnUrl,pkFields,newPk,table)
					});
				}
				break;
			case "update":
				if(watch.originData){
					if(watch.data){
						if(!_.isEqual(watch.originData,watch.data)){
							//记录被修改，但是可见范围没有变化
							sendSqlIDMessage(c,watch.sqlID,{
								opt:"update",
								originData:watch.originData,
								data:watch.data,
								btnUrl:buildBtnUrl(c,watch.watchSql.btnUrl,pkFields,newPk,table)
							});
						}
					}else{
						//原来有，现在不可见，作为删除发送
						sendSqlIDMessage(c,watch.sqlID,{opt:"updelete",originData:watch.originData});
					}
				}else{
					if(watch.data){
						//原来不可见,现在有，作为新增发送
						sendSqlIDMessage(c,watch.sqlID,{
							opt:"upinsert",
							data:watch.data,
							btnUrl:buildBtnUrl(c,watch.watchSql.btnUrl,pkFields,newPk,table)
						});
					}
				}
				break;
			case "delete":
				if(watch.originData){
					sendSqlIDMessage(c,watch.sqlID,{opt:"delete",originData:watch.originData});
				}
				break;
		}
	}
}
function convertFillValue(c,mdlname,colName,dataType,fillValue){
	if(fillValue===undefined){
		return undefined;
	}
	if(fillValue){
		try{
			return eval("("+fillValue+")");
		}catch(exception){
			throw fmt.Sprintf("convert model [%s] column [%s] fill value:[%s] error\n%s",mdlname,colName,fillValue,exception);
			//如果出错，则作为字符串返回
			//return fillValue;
		}
	}else{
		return null;
	}
}
function fillDefault(c,mdlname,fieldsets,columns,mainRow){
	//fill fieldsets
	for(var fieldName in fieldsets){
		for(var colIndex in columns){
			var column = columns[colIndex];
			if( column.Name == fieldName){
				column.Perm = fieldsets[fieldName].perm;
				var v = convertFillValue(c,mdlname,column.Name,column.DataType,fieldsets[fieldName].fill);
				if(v !== undefined){
					mainRow[column.Name]= v;
				}
				break;
			}
		}
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
			fillDefault(c,mdlname,fieldsets,args.model[mdlname].define.columns,args.model[mdlname].data[0]);
			args.SaveUrl = c.AuthUrl(url.SetQuery("mdlopt/save",{_n:c.GetTag("Element").name}));
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
			fillDefault(c,mdlname,fieldsets,args.model[mdlname].define.columns,args.model[mdlname].data[0]);
			args.SaveUrl = c.AuthUrl(url.SetQuery("mdlopt/save",{_n:c.GetTag("Element").name,_pk:pk}));
			break;
		}
		default:{
			throw fmt.Sprintf("the operate %q invalid!",operate);
		}
	}
	fireEvent(c,mdlname,"onRender",args);
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
	var _add = c.QueryValues()._add;
	var addition={}
	var Element = c.GetTag("Element");
	if(_add && _add.length>0){
		addition = c.Session.Get(_add[0]);
	}
	db.Open();
	try{
		mdlopt.FillByID(Element.name);
		if(mdlopt.RowCount() != 1){
			c.RenderError(fmt.Sprintf("the model operate:%q not found!",c.GetTag("Element").name));
		}else{
			var row = mdlopt.Row(0);
			var fieldsets = {};
			if(row.fieldsets){
				try{
					fieldsets = eval("("+row.fieldsets+")");
				}catch(exception){
					throw fmt.Sprintf("element [%s] fieldsets %s decode error\n%s",Element.name,row.fieldsets,exception);
				}
			}

			if(addition.fill){
				_.each(addition.fill,function(value,key){
					//如果没有自动填充的值，则从附加参数中取
					if(!fieldsets[key] || !fieldsets[key].fill){
						fieldsets[key]=fieldsets[key]||{};
						fieldsets[key].fill = value;
					}
				});
			}
			RenderMdlOpt(c,row.mdlname,row.operate,fieldsets,pk);
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
				//触发onSave事件
				fireEvent(c,row.mdlname,"onSave",{pk:pk,eleName:eleName,operate:row.operate,mdlname:row.mdlname});
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
