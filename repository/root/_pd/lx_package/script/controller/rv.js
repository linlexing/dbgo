var fmt = require("/fmt.js");
var url = require("/url.js");
var limit = 30;
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
			c.RenderRecordView({
				recordView:rvRow
			});
		}
	}
	finally{
		db.Close();
	}
}

exports.fetch=function(c){
	var eleName = c.TagPath;
	var db = c.Project.NewDBHelper();
	db.Open();
	try{
		var fetchOption = c.JsonBody;
		var rvTab = db.GetData("select datasrc,lab,col,pk,btn from lx_rv where name={{ph}}",eleName);
		var rev ={};
		if( rvTab.RowCount()>0){
			var rvTabRow = rvTab.Row(0);
			var rvLabel = rvTabRow.lab != "" ? eval("("+rvTabRow.lab+")") :{};
			var rvColumn = rvTabRow.col != "" ? eval("("+rvTabRow.col+")") :{};
			var rvPKFields = rvTabRow.pk.split(",");
			var rvBtn = eval("("+rvTabRow.btn+")");
			if(rvTabRow.datasrc!= ""){
				var tab = db.SelectLimitT(
					rvTabRow.datasrc,
					{
						CurrentGrade:c.CurrentGrade,
						UserName:c.UserName,
						UserDept:c.Session.Get("user.dept")
					},
					rvPKFields,
					fetchOption.lastkey,
					null,
					"",
					fetchOption.sort,
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
				rev.finish = tab.RowCount()<limit;
			}
		}else{
			rev.error="the sql is empty";
		}
		c.RenderJson(rev);
	}finally{
		db.Close();
	}
}
