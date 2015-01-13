app.controller('frmModelDataCtrl', ['$scope','$translate','$http','$alert',function ($scope,$translate,$http,$alert) {
	$scope.OriginData = angular.copy(
		_.object(_.map(G.mdlModel,function(val,key){
			return [key,val.data];
		}))
	);
	$scope.MainDefine = G.mdlModel[G.mdlOption.mdlname].define;
	$scope.MainRow = G.mdlModel[G.mdlOption.mdlname].data[0];
	$scope.Option = G.mdlOption;
	$scope.MainRowIsUnchanged = function(){
		return angular.equals($scope.MainRow ,$scope.OriginData[G.mdlOption.mdlname][0]);
	}
	$scope.getColumn=function(colName){
		for(var i in $scope.MainDefine.columns){
			var col = $scope.MainDefine.columns[i];
			if(col.Name == colName){
				return col;
			}
		}
		throw "can't find column:"+colName;
	}
	$scope.getColumnLabel=function(colName){
		for(var i in $scope.MainDefine.columns){
			var col = $scope.MainDefine.columns[i];
			if(col.Name == colName){
				if(!col.Desc || !col.Desc.Label){
					return col.Name;
				}
				if( col.Desc.Label[$translate.use()]){
					return col.Desc.Label[$translate.use()];
				}
				return col.Desc.Label.en||col.Desc.Label.cn;
			}
		}
		return "";
	}
	$scope.getRequired=function(colName){
		return _.contains($scope.MainDefine.pk,colName);
	}
	$scope.getReadonly=function(colName){
		return Option.operate=='delete' || Option.operate=='browse'||$scope.getColumn(colName).Perm=='readonly';
	}
	$scope.save = function(){
		$scope.mdlData = {
			originData : $scope.OriginData,
			data:{}
		};
		$scope.$broadcast("model.save");
		$http.post(G.mdlSaveUrl,$scope.mdlData)
			.success(function(data,status,headers,config,statusText){
				if(data.ok){
					$scope.close();
				}else{
					$alert({title: 'error', content:data&&data.error?data.error:("<textarea rows='15' cols='80' wrap='off' readonly class='err-textarea'>" +data+"</textarea>")||statusText, placement: 'top-right', type: 'danger', show: true});
				}

			})
			.error(function(data,status,headers,config,statusText){
				$alert({title: 'error', content:data&&data.error?data.error:statusText, placement: 'top-right', type: 'danger', show: true});
			});
	}
	$scope.close=function(){
		close();
	}
	$scope.$on("model.save",function(targetScope){
		if($scope.Option.operate != "delete"){
			$scope.mdlData.data[$scope.Option.mdlname] = [$scope.MainRow];
		}
	});
}]);

app.directive("lxField", ["$compile","$translate",function ($compile,$translate) {
	return {
	    restrict: "E",
		scope:{
			fieldName:"=",
		},
		replace:true,
		template:[
			"<div class='form-group' ng-show=\"define.Perm!='hidden'\">",
				"<label class='col-sm-2 control-label'>{{fieldLabel()}}:</label>",
				"<div class='col-sm-10'>",
					"<input type='text' class='form-control' ng-required='required' ng-model=\"record[fieldName]\" ng-readonly='readonly' ></input>",
				"</div>",
			"</div>"].join(""),
		link:function(scope,element,attris){
			scope.define = scope.$parent.getColumn(scope.fieldName);
			scope.required = scope.$parent.getRequired(scope.fieldName);
			scope.readonly = scope.$parent.getReadonly(scope.fieldName);
			scope.fieldLabel = function(){
				if(!scope.define.Desc || !scope.define.Desc.Label){
					return scope.define.Name;
				}
				if( scope.define.Desc.Label[$translate.use()]){
					return scope.define.Desc.Label[$translate.use()];
				}
				return scope.define.Desc.Label.en||scope.define.Desc.Label.cn;
			}
			scope.record = scope.$parent.MainRow;
		}
	};
}]);


