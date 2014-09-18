/*the data struct is :
	op:		t-table u-update d-delete i-insert
	tab:	when op is t,this var is table
	pkSql:	when op not is t,this store the pk values select sql,the result column name must equ table's pk column name
	sets:	when op is u,this store set express,e.g. {col1:"'a'",col2:"12+12"}
*/
exports.do=function(db,tabName,data){

}
