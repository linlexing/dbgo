var fmt = require("/fmt.js");
var limit = 30;
exports.show=function(c){
	var rv=c.DBModel("lx_rv")[0];
	var db =rv.DBHelper();
	db.Open();
	try{
		rv.FillByID(c.GetTag("Element").name);
	}
	finally{
		db.Close();
	}
	if(rv.RowCount()==0){
		c.RenderError(fmt.Sprintf("the lx_rv can't find name=%q record.",c.GetTag("Element").name));
	}else{
		var rvRow = rv.Row(0);
		c.RenderRecordView({
			recordView:rvRow
		});
	}
}

exports.fetch=function(c){
	var eleName = c.TagPath;
	var db = c.Project.NewDBHelper();
	db.Open();
	try{
		var fetchOption = c.JsonBody;
		var rvTab = db.GetData("select datasrc,lab,col,pk from lx_rv where name={{ph}}",eleName);
		var rev ={};
		if( rvTab.RowCount()>0){
			var rvTabRow = rvTab.Row(0);
			var rvLabel = rvTabRow.lab != "" ? eval("("+rvTabRow.lab+")") :{};
			var rvColumn = rvTabRow.col != "" ? eval("("+rvTabRow.col+")") :{};
			var rvPKFields = rvTabRow.pk.split(",");
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
					null,
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
				rev.data = tab.Rows();
			}
		}else{
			rev.error="the sql is empty";
		}
		c.RenderJson(rev);
	}finally{
		db.Close();
	}
}
