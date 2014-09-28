var fmt = require("/fmt.js");
var url = require("/url.js");
var uuid = require("/uuid.js");
var sha256 = require("/sha256.js");
var convert = require("/convert.js");
var cache = require("watch/cache.js");
var maxLimit = 200;
function getElementUrl(db,eleName){
	return db.QStr("select url from lx_element where name={{ph}}",eleName);
}

exports.show=function(c){
	var rv=c.DBModel("lx_rv")[0];
	var db =rv.DBHelper();
	db.Open();
	try{
		rv.FillByID(c.GetTag("Element").name);
		if(rv.RowCount()==0){
			c.RenderError(fmt.Sprintf("the lx_rv can't find name=%q record.",c.GetTag("Element").name));
		}else{
			var rvRow = rv.Row(0);
			var rvBtn = eval("("+rvRow.btn+")");
			for(var i in rvBtn){
				if(!rvBtn[i].bindrecord){
					var eleUrl = getElementUrl(db,rvBtn[i].elename);

					if(eleUrl){
						rvBtn[i].url=c.AuthUrl(url.SetQuery(eleUrl,{_ele:rvBtn[i].elename}));
					}
				}
			}
			rvRow.btn =rvBtn;
			rvRow.uuid = uuid.NewRandom();
			c.RenderRecordView({
				recordView:rvRow
			});
		}
	}
	finally{
		db.Close();
	}
}
function getWatchSql(db,dataSrc,templateParam,pkFields,lastKey,columns,sort){
	var where  =[];
	for(var i in pkFields){
		where.push(fmt.Sprintf("%s = {{ph}}",pkFields[i]));
	}
	var whereStr = "";
	if(where.length>0){
		whereStr = where.join(" AND ");
	}
	var rev =db.BuildSelectLimitSql(dataSrc,pkFields,lastKey,columns,whereStr,sort,1);
	rev.sql = db.ConvertSql(rev.sql,templateParam);
	return rev;
}
exports.fetch=function(c){

	var eleName = c.TagPath;
	var db = c.Project.NewDBHelper();
	db.Open();
	try{
		var fetchOption = c.JsonBody;
		var rvTab = db.GetData("select datasrc,lab,col,pk,btn,basetable from lx_rv where name={{ph}}",eleName);
		var rev ={};
		if( rvTab.RowCount()>0){
			var rvTabRow = rvTab.Row(0);
			var rvLabel = rvTabRow.lab != "" ? eval("("+rvTabRow.lab+")") :{};
			var rvColumn = rvTabRow.col != "" ? eval("("+rvTabRow.col+")") :{};
			var rvPKFields = rvTabRow.pk.split(",");
			var rvBtn = eval("("+rvTabRow.btn+")");
			var limit = Math.min(fetchOption.limit,maxLimit);
			var templateParam ={
						CurrentGrade:c.CurrentGrade,
						UserName:c.UserName,
						UserDept:c.Session.Get("user.dept")
			};
			var sortStrArray=[];
			var descSortStrArray=[];
			if( fetchOption.sort.length>0){
				for(var i in fetchOption.sort){
					if(fetchOption.sort[i].type == "DESC"){
						sortStrArray.push(fetchOption.sort[i].column + " DESC");
						descSortStrArray.push(fetchOption.sort[i].column);
					}else{
						sortStrArray.push(fetchOption.sort[i].column);
						descSortStrArray.push(fetchOption.sort[i].column + " DESC");
					}
				}
			}else{
				//如果没有排序，则是主键值的倒序
				for(var i in rvPKFields){
					descSortStrArray.push(rvPKFields[i] + " DESC");
				}
			}

			if(rvTabRow.datasrc!= ""){
				var tab = db.SelectLimitT(
					rvTabRow.datasrc,
					templateParam,
					rvPKFields,
					fetchOption.lastKey,
					null,
					"",
					sortStrArray,
					limit
				)
				rev.columns = [];
				var tabColumns = tab.Columns();
				for(var colName in tabColumns){
					rev.columns.push(
						{
							fieldName:colName,
							displayName:{
								en:rvLabel[colName]?rvLabel[colName].en||colName: colName ,
								cn:rvLabel[colName]?rvLabel[colName].cn||colName: colName
							}
						}
					);
				}
				rev.btnUrl = [];
				for(var iRowIndex = 0 ;iRowIndex < tab.RowCount();iRowIndex++){
					var iCount = 0;
					var oneUrlArr =[];
					for(var i in rvBtn){
						if(rvBtn[i].bindrecord){
							var btnUrl = null;
							var eleUrl = getElementUrl(db,rvBtn[i].elename);
							if(eleUrl){
								btnUrl = c.AuthUrl(url.SetQuery(eleUrl,{_ele:rvBtn[i].elename,_pk:tab.GetStrings(iRowIndex,rvPKFields)}));
							}
							oneUrlArr.push(btnUrl);
						}
					}
					rev.btnUrl.push(oneUrlArr);
				}
				rev.data = tab.Rows();
				rev.first = !fetchOption.lastKey;
				rev.finish = tab.RowCount()<limit;
				//watch basetable
				if( rvTabRow.basetable ){
					var endKeyValues;
					if(tab.RowCount() >0){
						endKeyValues = _.pick(
							tab.Row(tab.RowCount()-1),
							_.pluck(fetchOption.sort,"column"),
							rvPKFields
						)
					}
					var watchSql = getWatchSql(db,rvTabRow.datasrc,templateParam,rvPKFields,endKeyValues,null,descSortStrArray);
					var sqlID = convert.EncodeBase64(sha256.Sum256(convert.Str2Bytes(watchSql.sql+JSON.stringify(watchSql.params))));
					cache.Put(c,rvTabRow.basetable,watchSql,fetchOption.uuid);
				}
			}
		}else{
			rev.error="the sql is empty";
		}
		c.RenderJson(rev);
	}finally{
		db.Close();
	}
}
