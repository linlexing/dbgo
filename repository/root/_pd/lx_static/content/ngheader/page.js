//PGType const
var i =0;
var TypeString 			= i++;
var TypeBool 			= i++;
var TypeInt64 			= i++;
var TypeFloat64 		= i++;
var TypeTime 			= i++;
var TypeBytea 			= i++;
var TypeStringSlice 	= i++;
var TypeBoolSlice 		= i++;
var TypeInt64Slice	 	= i++;
var TypeFloat64Slice 	= i++;
var TypeTimeSlice 		= i++;
var TypeJSON 			= i++;
var TypeJSONSlice 		= i++;
//check.level
i = 0;
var CHECK_LEVEL_DISABLE = i++;//禁用
var CHECK_LEVEL_ACCEPT  = i++;//出错时可以保存
var CHECK_LEVEL_FORCE   = i++;//出错时可以强制保存
var CHECK_LEVEL_REFUSED = i++;//出错时不能保存，也不能强制保存
//Bill Operate style
i = 0;
var BILL_ADD 	= i++;
var BILL_EDIT	= i++;
var BILL_DELETE = i++;
var BILL_BROWSE = i++;
function regexp_like(value,regstr){
	return new RegExp(regstr,"m").test(value);
}
var app = angular.module('app',appDependencys);
app.run(['$rootScope','$log','$window','$alert', function ($rootScope, $log, $window,$alert) {
	$rootScope.$log = $log;
	$rootScope.viewport ='';
	$rootScope._ = _;
	$rootScope.mediaquery = function(){
		$rootScope.$apply(function(){
			if ($(".bitdb-view-xs").css("display") == "block" ){
				$rootScope.viewport = 'xs';
			}
			else if ($(".bitdb-view-sm").css("display") == "block" ){
				$rootScope.viewport = 'sm';
			}
			else if ($(".bitdb-view-md").css("display") == "block" ){
				$rootScope.viewport = 'md';
			}
			else if ($(".bitdb-view-lg").css("display") == "block" ){
				$rootScope.viewport = 'lg';
			}
		});
		return $rootScope.viewport;
	}
	$rootScope.mediaquery();
	angular.element($window).bind('resize',$rootScope.mediaquery);
	$rootScope.clearMessage = function(){
		if(alertMessage){
			alertMessage.hide();
		}
	}
	var alertMessage;
	$rootScope.gotoMessage = function(mes){
		alertMessage = $alert({title: '', content:mes, placement: 'top', type: 'info', show: true});
		console.log(alertMessage);
		$rootScope.finishedMessage = true;
	}
}]);
app.config(['$translateProvider', function ($translateProvider) {
	if( G.translate_en){
		$translateProvider.translations('en', G.translate_en);
	}
	if( G.translate_cn){
		$translateProvider.translations('cn', G.translate_cn);
	}
	$translateProvider.registerAvailableLanguageKeys(['en', 'cn'], {
		'en_US': 'en',
		'en_UK': 'en',
		'zh_CN': 'cn',
		'zh_cn': 'cn'
	}).determinePreferredLanguage()
	.useLocalStorage();
}]);
app.run(["$translate","$rootScope",function ($translate,$rootScope) {
	$translate("Title").then(function(t){
		document.title=G.projectName+" - " + t;
	});
	$rootScope.$on('$translateChangeSuccess', function () {
		$translate("Title").then(function(t){
			document.title=G.projectName+" - " + t;
		});
	});
}]);
