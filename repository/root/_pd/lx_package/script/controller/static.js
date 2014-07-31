exports.Public=true;
exports.file=function(c){
  c.RenderStaticFile(c.TagPath);
}
exports.userfile=function(c){
  c.RenderUserFile(c.TagPath);
}
