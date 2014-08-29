var spinner;
function showSpin(){
	var opts = {
	  lines: 8, // The number of lines to draw
	  length: 2, // The length of each line
	  width: 2, // The line thickness
	  radius: 4, // The radius of the inner circle
	  corners: 1, // Corner roundness (0..1)
	  rotate: 0, // The rotation offset
	  direction: 1, // 1: clockwise, -1: counterclockwise
	  color: '#000', // #rgb or #rrggbb or array of colors
	  speed: 1, // Rounds per second
	  trail: 60, // Afterglow percentage
	  shadow: false, // Whether to render a shadow
	  hwaccel: false, // Whether to use hardware acceleration
	  className: 'spinner', // The CSS class to assign to the spinner
	  zIndex: 2e9, // The z-index (defaults to 2000000000)
	  top: '50%', // Top position relative to parent
	  left: '50%' // Left position relative to parent
	};
	spinner = new Spinner(opts).spin($("#spinDiv")[0]);

}
function hideSpin(){
	if(spinner)
		spinner.stop();
}
window.WebSocket = window.WebSocket || window.MozWebSocket;
app.controller('mainCtrl', ['$translate', '$scope','$alert','$http','$timeout',function ($translate, $scope,$alert,$http,$timeout) {
	var fetchOption={};
	$translate("LABEL").then(function(label){
		document.title=label;
	});
	$scope.pending = 0;
	$scope.time = 10.0;
	$scope.search = {field :null,opt:"equ",value:""};
	$scope.sort = {};
	$scope.searchOpts = ["OPT_PREX","OPT_EQ","OPT_LIKE","OPT_NE","OPT_LT","OPT_LE","OPT_GT","OPT_GE","OPT_IN","OPT_NIN","OPT_REGEXP","OPT_SUFX"];
	$scope.navCollapsed=true;
	$scope.data={columns:[],data:[],total:0};
	$scope.scrollTop = 0;
	$scope.$translate = $translate;
	$(window).scroll(function (e){
		var top = $(window).scrollTop();
		$scope.$apply(function(){
			$scope.thStyle ={top:Math.round(Math.max(0,top-$(".content table").offset().top))+"px"} ;
		});
		if ($(window).scrollTop() >0 && $(window).scrollTop() + $(window).height() >= $(document).height()){
			$scope.fetchData();
		}
	});
	$scope.$watch("pending",function(newvalue,oldvalue){
		if(newvalue >0){
			showSpin();
		}else{
			hideSpin();
		}
	});
	$scope.onMessage=function(evt){
		var data = eval("("+evt.data+")");
		$scope.$apply(function(){
			$scope.fetchInfo.time = new Date()-$scope.fetchInfo.startTime;
			$scope.fetchInfo.count = data.data.length;
			var No = $scope.data.data.length;
			$scope.data.data = $scope.data.data.concat(
				_.map(data.data,function(value,i){
					return _.extend(value,{_No_:++No});
				})
			);
			if($scope.data.columns.length==0){
				$scope.data.columns = data.columns;
			}
			$scope.pending --;
		});
	}
    var websocket ;
	if (window.WebSocket){
	    websocket = new WebSocket("wss://" + window.location.host +G.rv_watchAction);
	    websocket.onopen = function(evt){
			//$scope.refreshData();
		};
	    websocket.onclose =  function(evt){
			console.log("close\n");
			console.log(evt);
		};
	    websocket.onmessage = $scope.onMessage;
	}
	$scope.fetchData=function(){
		if($scope.data.data.length>0){
			fetchOption.lastkey = _.map(
				_.pick(_.last($scope.data.data),$scope.sort.column,G.rv_primaryKeys.split(",")),
				function(value,key){
					return {field:key,value:value};
				}
			);
		}
		$scope.fetchInfo = {
			startTime:new Date(),
		}
		$scope.pending ++;
		//websocket.send(JSON.stringify({event:"rv_fetchData",data:fetchOption})) ;
		$http.post(G.rv_dataAction,fetchOption)
			.success(function(data,status,headers,config,statusText){
				try{
					$scope.fetchInfo.time = new Date()-$scope.fetchInfo.startTime;
					$scope.fetchInfo.count = data.data.length;
					var No = $scope.data.data.length;
					$scope.data.data = $scope.data.data.concat(
						_.map(data.data,function(value,i){
							return _.extend(value,{_No_:++No});
						})
					);
					if($scope.data.columns.length==0){
						$scope.data.columns = data.columns;
					}
				}finally{
					$scope.pending --;
				}
			})
			.error(function(data,status,headers,config,statusText){
				$alert({title: 'error', content:data&&data.error?data.error:statusText, placement: 'top-right', type: 'danger', show: true});
				$scope.fetchInfo.time = new Date()-$scope.fetchInfo.startTime;
				$scope.pending --;
			});

	}
	$scope.refreshData= function(){
		$scope.navCollapsed = true;
		fetchOption ={
			search:{
				field:$scope.search.field ? $scope.search.field.fieldName :"",
				opt:$scope.search.opt,
				value:$scope.search.value
			},
			sort:$scope.sort
		};
		$scope.data = {columns:[],data:[],total:0};
		$scope.fetchData();
	}
	$scope.thClick=function(col){
		if($scope.pending!=0){
			return;
		}
		if($scope.sort.column && $scope.sort.column == col.fieldName && $scope.sort.type == "DESC"){
			$scope.sort.type = "ASC";
		}else if($scope.sort.column && $scope.sort.column == col.fieldName && $scope.sort.type == "ASC"){
			$scope.sort = {};
		}else{
			$scope.sort.column = col.fieldName;
			$scope.sort.type = "DESC";
		}
		$scope.refreshData();
	}
	$scope.refreshData();
}]);
