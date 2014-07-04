exports.clearcache=function(c){
  c.Project.ClearCache();
  c.RenderJson({ok:1});
}