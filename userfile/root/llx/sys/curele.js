
function CurrentUserElement(){
	data = [["changepwd","root","01.Change Password","01.修改密码","90.System/10.Rights","90.系统管理/10.权限管理","/root/changepwd/show?_ele=changepwd"],["home_default","root","welcome","欢迎","10.Home/10.Welcome","10.首页/10.欢迎","/root/home/default?_ele=home_default"],["rolemanager","root","01.Role Manager","01.角色管理","80.Define/10.Core","80.系统定制/10.核心模块","/root/rv/show?_ele=rolemanager"],["test1","root","11.11Role Manager","11.11角色管理","80.Define/10.Core","80.系统定制/10.核心模块",null],["test2","root","05.05Role Manager","01.05角色管理","80.Define/10.Core","80.系统定制/10.核心模块",null],["test3","root","03.03Role Manager",null,null,"80.系统定制/10.核心模块",null],["test4","root",null,"02.测试1",null,"80.系统定制/11.核心模块1",null],["test5","root",null,"02.测试2",null,"80.系统定制/11.核心模块1/20.测试",null],["test6","root",null,"02.测试3",null,"80.系统定制/11.核心模块1/20.测试",null]];
	cols = ["name","grade","label_en","label_cn","category_en","category_cn","url"];
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