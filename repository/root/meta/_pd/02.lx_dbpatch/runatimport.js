function(dbHelper,tmpTable,exportGrade,sqlWhere){
	tab = dbHelper.GetData("select script from "+ tmpTable);
	for(i=0;i < tab.RowCount();i++){
		row = tab.Row(i);
		dbHelper.GoExecT(row.script,{
			TmpTableName:tmpTable,
			ExportGrade:exportGrade,
			SqlWhere:sqlWhere
		});
	}
}
