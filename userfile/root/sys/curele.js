
function CurrentUserElement(){
	data = [];
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