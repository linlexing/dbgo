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
var app = angular.module('app',['ngAnimate','mgcrea.ngStrap','pascalprecht.translate']);
app.directive('lxField', ['$compile',function ($compile) {
  return {
    restrict: 'A',
    link: function (scope, element,attrs) {
	  if(!attrs.ngModel){
        attrs.$set("ngModel",attrs.lxField);
	    $compile(element)(scope);
	  }
    }
  };
}]);
