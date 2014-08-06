
function CurrentUserElement(){
	data = [["changepwd","root","01.Change Password","90.System/10.Rights","changepwd/show","01.修改密码","90.系统管理/10.权限管理"],["rolemanager","root","01.Role Manager","80.Define/10.Core",null,"01.角色管理","80.系统定制/10.核心模块"],["test1","root","11.11Role Manager","80.Define/10.Core",null,"11.11角色管理","80.系统定制/10.核心模块"],["test2","root","05.05Role Manager","80.Define/10.Core",null,"01.05角色管理","80.系统定制/10.核心模块"],["test3","root","03.03Role Manager",null,null,null,"80.系统定制/10.核心模块"],["test4","root",null,null,null,"02.测试1","80.系统定制/11.核心模块1"],["test5","root",null,null,null,"02.测试2","80.系统定制/11.核心模块1/20.测试"],["test6","root",null,null,null,"02.测试3","80.系统定制/11.核心模块1/20.测试"]];
	cols = ["name","grade","label_en","category_en","url","label_cn","category_cn"];
	rev = [];
	for(rowIndex in data){
		rowObject = {};
		for(colIndex in cols){
			rowObject[cols[colIndex]] = data[rowIndex][colIndex];
		}
		rev.push(rowObject);
	}
	return rev;
}