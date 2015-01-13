
app.controller('termController', ['$scope','$translate','$http','$alert',function ($scope,$translate,$http,$alert) {
	function checkTerms(value){
		if(value&&(value.length==0||_.last(value).column)){
			value.push({
				logic:"And"
			});
		}
	}
	$scope.terms =[];
	$scope.searchOpts = ["OPT_PREX","OPT_EQ","OPT_LIKE","OPT_NE","OPT_LT","OPT_LE","OPT_GT","OPT_GE","OPT_IN","OPT_NIN","OPT_REGEXP","OPT_SUFX"];

	if($scope.MainRow.term){
	 	$scope.terms = eval("("+$scope.MainRow.term+")");
	}
	/*terms struct is :
	[
		{
			left:"(",
			column:"name",
			operate:"OPT_EQ",
			value:"abc",
			right:")",
			logic:"AND",
		},
		...
	]
	*/
	checkTerms($scope.terms);
	$scope.$watch("terms",function(newValue,oldValue){
		checkTerms(newValue);
	},true);
	$scope.termColumns = G.rv_info.columns.concat();
	$scope.selectColumns=_.map(G.rv_info.columns,function(value){
		return {
			name:value.name,
			label:value.label,
			pk:_.contains(G.rv_info.pk,value),
			selected:false,
			hidden:false
		}
	});
	$scope.getColumnLabel=function(column){
		if(column.label&&column.label[$translate.use()]){
			return column.label[$translate.use()];
		}else{
			return column.name;
		}
	}
	$scope.termAdd = function(idx){
		$scope.terms.splice(idx,0,{logic:"And"});
	}
	$scope.termDelete = function(idx){
		$scope.terms.splice(idx,1);
	}
}]);

