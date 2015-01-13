exports.SignQuoted=function(val){
	var rev = "'";
	for(var i = 0;i < val.length;i++){
		c = val.charAt(i);
		switch(c){
			case "'":
				rev += "''";
				break;
			case "\n":
				rev += "\\n";
				break;
			case "\r":
				rev += "\\r";
				break;
			default:
				rev += c;
				break;
		}
	}
	rev += "'";
	return rev;
}
