exports.opt=function(c){
  pArray = c.TagPath.split("/");
  optName = pArray[0];
  var keys
  if(pArray.length > 1 ){
    keys = _.map(pArray.slice(1),function(val){
      //convert to array if is '[..]'
      if(val.length > 1&& val[0]=="[" && val[val.length-1]=="]"){
        return eval(val);
      }else{
        return val;
      }
    });
  }
  c.RenderModelOperate(optName,keys);
}