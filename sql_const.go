package main

const (
	SQL_GerVersion    = "select grade,max(verno) as verno from lx_version group by grade"
	SQL_GetIntercept  = "select id,whenint,script,grade from lx_intercept order by id"
	SQL_GetController = "select script,public,grade from lx_controller where name=$1"
	SQL_GetCheck      = `
		select
			a.id,
			a.displaylabel,
			a.level,
			a.fields||b.fields as fields,
			a.runatserver or b.runatserver as runatserver,
			array_to_string(array['('||b.script||')','('||a.script||')'],'&&') as script,
			array_to_string(array['('||b.sqlwhere||')','('||a.sqlwhere||')'],' AND ') as sqlwhere,
			a.grade
		from lx_check a left join lx_checkaddition b on a.tablename=b.tablename and a.addition=b.addition
		where a.tablename=$1`
	SQL_GetAction       = "select id,script,grade from lx_action where ctrlname=$1 order by id"
	SQL_GetView         = "select name,content,grade from lx_view order by name"
	SQL_GetCheckResult  = "select pks,checkid,grade,refreshtime,refreshby from lx_checkresult where tablename=$1 and pks=$2 and $3 like grade||'%'"
	SQL_GetPackage      = `select script from lx_package where filename like $1||'%' and right(filename,length(filename)-length($1))~'^[^/]*$' and grade_canuse($2,grade) order by filename`
	SQL_GetPackageNames = `select filename from lx_package where filename like $1||'%' and right(filename,length(filename)-length($1))~'^/?[^/]+\.js$' and grade_canuse($2,grade) order by filename`
	SQL_GetStatic       = "select filename,content,lasttime from lx_static order by filename"
)
