function(dbHelper,tmpTable,exportGrade,sqlWhere){
	tab = dbHelper.GetData("select script from "+ tmpTable);
	for(row in tab.Rows()){
		dbHelper.ExecT(row.script,{
			TmpTableName:tmpTable,
			ExportGrade:exportGrade,
			SqlWhere:sqlWhere
		});
	}
}
