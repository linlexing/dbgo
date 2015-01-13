var spinner;
var recordLimit = Math.round($(window).height()/40*3);

/**
 * Find insertion point for a value val, as specified by the comparator
 * (a function)
 * @param sortedArr The sorted array
 * @param val The value for which to find an insertion point (index) in the array
 * @param comparator The comparator function to compare two values
 */
function findInsertionPoint(sortedArr, val, comparator) {
   var low = 0, high = sortedArr.length;
   var mid = -1, c = 0;
   while(low < high)   {
      mid = parseInt((low + high)/2);
      c = comparator(sortedArr[mid], val);
      if(c < 0)   {
         low = mid + 1;
      }else if(c > 0) {
         high = mid;
      }else {
         return mid;
      }
      //alert("mid=" + mid + ", c=" + c + ", low=" + low + ", high=" + high);
   }
   return low;
}
function cmpValue(v1,v2){
	if(typeof v1.getMonth === 'function'){
		return v1.valueOf()- v2.valueOf();
	}else{
		if(v1 == v2){
			return 0
		}
		if(v1 < v2){
			return -1;
		}
		return 1;
	}
}
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
app.controller('mainCtrl', ['$popover','$translate', '$scope','$alert','$http','$timeout',function ($popover,$translate, $scope,$alert,$http,$timeout) {
	function createWS(){
	    var websocket ;
		if (window.WebSocket){
			websocket = new ReconnectingWebSocket("wss://" + window.location.host +G.rv_watchAction);
		    websocket.onopen = function(evt){
  				$scope.$apply(function(){
					$scope.offline=false;
				});
			};
		    websocket.onclose =  function(evt){
				$scope.$apply(function(){
					$scope.offline = true;
				});
			};
			websocket.onerror=function(ev){
				console.log(ev);
			};
		    websocket.onmessage = function(evt){
				if(evt.data ){
					var evData = eval("("+evt.data+")");
					switch(evData.opt){
						case "insert":
						case "upinsert":
							$scope.$apply(function(){
								$scope.insertRow(evData.data,evData.btnUrl);
							});
							break;
						case "update":
							$scope.$apply(function(){
								$scope.deleteRow(evData.originData);
								$scope.insertRow(evData.data,evData.btnUrl);
							});
							break;
						case "delete":
						case "updelete":
							$scope.$apply(function(){
								$scope.deleteRow(evData.originData);
							});
							break;
						default:
							throw "opt:"+evData.opt + " invalid";
					}
				}else{
					console.log("websocket onmessage:");
					console.log(evt);
				}
			}
		}
	}
	var fetchOption;
	$translate("LABEL").then(function(label){
		document.title=label;
	});
	$scope.thStyle = null;
	$scope.navbarStyle = null;
	$scope.data = {columns:[],data:[],total:-1,finish:false,btnUrl:[]};
	$scope.selected = null;
	var recordBtnPop = null;
	$scope.deleteRow = function(rowData){
		var pk = $scope.define.pk.split(",");
		for(var i in $scope.data.data){
			var equ = true;
			for(var j in pk){
				if(cmpValue($scope.data.data[i][pk[j]],rowData[pk[j]])!=0){
					equ=false;
					break;
				}
			}
			if(equ){
				$scope.data.data.splice(i,1);
				$scope.data.btnUrl.splice(i,1);
				break;
			}
		}
	}
	$scope.insertRow = function(rowData,btnUrl){
		var pk = $scope.define.pk.split(",");
		var sort = $scope.sort;
		for(var i in sort){
			var pkIndex = _.indexOf(pk,sort[i].column);
			if(pkIndex > -1){
				pk.splice(pkIndex,1);
			}
		}
		//add pk column if not include
		for(var i in pk){
			sort.push({column:pk[i]});
		}
		console.log("sort:",sort);
		var insertIndex = findInsertionPoint($scope.data.data,rowData,function(v1,v2){
			for(var i in sort){
				var colName = sort[i].column;
				var cmp = cmpValue(v1[colName],v2[colName]);
				if(sort[i].type=="DESC"){
					cmp = - cmp;
				}
				if(cmp != 0){
					return cmp;
				}
			}
			return 0;
		});
		$scope.data.data.splice(insertIndex,0,rowData);
		$scope.data.btnUrl.splice(insertIndex,0,btnUrl);
	}
	$scope.selectedRow = function(row,ev){
		if($scope.selected == row){
			recordBtnPop.toggle();
			return;
		}
		if(recordBtnPop){
			recordBtnPop.hide();
		}
		$scope.selected = row;
		recordBtnPop = $popover($(ev.target),{animation:"am-fade-and-scale",template:"recordBtnTemplate.html",trigger:'manual',container:'#mainCtrl',placement:"top"});
		var idx = 0;
		var btns = [];
		for(var i in $scope.define.btn){
			if($scope.define.btn[i].bindrecord){
				var v = _.clone($scope.define.btn[i]);
				v.url = $scope.data.btnUrl[$scope.data.data.indexOf(row)][idx++];
				btns.push(v);
			}
		}
		recordBtnPop.$scope.pkValues = _.values(_.pick(row,$scope.define.pk.split(","))).join(",");
		recordBtnPop.$scope.btn = btns;
		recordBtnPop.$scope.$translate=$translate;
		recordBtnPop.$promise.then(recordBtnPop.show);
	}
	$scope.selectedIndex=function(){
		return $scope.data.data.indexOf($scope.selected);
	}
	$scope.pending = 0;
	$scope.time = 10.0;
	$scope.search = {field :null,opt:"equ",value:""};
	$scope.sort = [];
	$scope.searchOpts = ["OPT_PREX","OPT_EQ","OPT_LIKE","OPT_NE","OPT_LT","OPT_LE","OPT_GT","OPT_GE","OPT_IN","OPT_NIN","OPT_REGEXP","OPT_SUFX"];
	$scope.navCollapsed=true;
	$scope.define = G.rv_define;
	$scope.scrollTop = 0;
	$scope.$translate = $translate;
	$("#divHScroll").scroll(function(e){
		var left = $("#divHScroll").scrollLeft();
		$(".content").scrollLeft(left);
	});
	$(window).scroll(function (e){
		var top = $(window).scrollTop();
		$scope.$apply(function(){
			$scope.thStyle ={top:Math.round(Math.max(0,
				top-$("#mainCtrl").offset().top-$(".fetchinfo").height()
				))+"px"} ;
			$scope.navbarStyle ={top:Math.round(Math.max(0,top-$("#mainCtrl").offset().top))+"px"} ;
		});
		if ($(window).scrollTop() >0 && $(window).scrollTop() + $(window).height() >= $(document).height()){
			if(!$scope.data.finish){
				$scope.fetchData(_.pick(
						_.last($scope.data.data),
						_.pluck($scope.sort,"column"),
						$scope.define.pk.split(",")
					),$scope.sort,Math.round(recordLimit/2));
			}
		}
	});
	$scope.$watch("pending",function(newvalue,oldvalue){
		if(newvalue >0){
			showSpin();
		}else{
			hideSpin();
		}
	});
	$scope.getSort=function(colName){
		for(var i in $scope.sort){
			if($scope.sort[i].column == colName){
				return $scope.sort[i];
			}
		}
		return null;
	}

	createWS();
	$scope.getItemPK=function(item){
		return _.values(_.pick(item,$scope.define.pk.split(","))).join(",");
	}
	$scope.fetchData=function(lastKey,sort,limit){
		var ops = {
			lastKey:lastKey,
			limit:limit,
			uuid:$scope.define.uuid
		}
		$scope.fetchInfo = {
			startTime:new Date(),
		}
		ops.sort=sort;
		$scope.pending ++;
		//websocket.send(JSON.stringify({event:"rv_fetchData",data:fetchOption})) ;
		$http.post(G.rv_dataAction,ops)
			.success(function(data,status,headers,config,statusText){
				try{
					if(typeof data == "string"){
						$alert({title: 'error', content:("<textarea rows='15' cols='80' wrap='off' readonly class='err-textarea'>" +data+"</textarea>")||statusText, placement: 'top-right', type: 'danger', show: true});
						return;
					}
					$scope.fetchInfo.time = new Date()-$scope.fetchInfo.startTime;
					$scope.fetchInfo.count = data.data.length;
					if(!lastKey){
						console.log(data);
						$scope.data.columns = data.columns;
						$scope.data.data = [];
						$scope.data.btnUrl = [];
						$(window).scrollTop(0);
					}
					//add new record
					for(var i in data.data){
						$scope.data.data.push(data.data[i]);
						$scope.data.btnUrl.push(data.btnUrl[i]);
					}
					$scope.data.finish = data.finish;
					if($scope.data.finish){
						$scope.data.total=$scope.data.data.length;
					}
					$timeout(function(){
						$("#divHScroll .data").width($("#tabData").width());
					},0,false);

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
		if(recordBtnPop){
			recordBtnPop.hide();
			recordBtnPop=null;
			$scope.selected = null;
		}
		$scope.fetchData(null,$scope.sort,recordLimit);
	}
	$scope.thClick=function(col){
		if($scope.pending!=0){
			return;
		}
		if($scope.sort.length ==1){
			var sort = $scope.sort[0];
			if (sort.column && sort.column == col.fieldName){
				if(sort.type == "DESC"){
					sort.type = "ASC";
				}else if(sort.type == "ASC"){
					$scope.sort = [];
				}
			}else{
				$scope.sort = [{column:col.fieldName,type:"DESC"}];
			}
		}else{
			$scope.sort = [{column:col.fieldName,type:"DESC"}];
		}
		$scope.refreshData();
	}
	$scope.refreshData();
}]);
