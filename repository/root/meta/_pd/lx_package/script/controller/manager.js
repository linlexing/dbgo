exports.Public = true;
exports.clearcache=function(c){
  c.Project.ClearCache();
  c.RenderJson({ok:1});
}
exports.reload=function(c){
  c.Project.ClearCache();
  c.Project.ReloadRepository();
  c.RenderJson({ok:1});
}
