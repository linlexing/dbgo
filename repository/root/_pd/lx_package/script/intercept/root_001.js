var fmt= require("/fmt.js");
var cnst = require("/const.js");
exports.When =cnst.BFORE;
exports.Intercept = function(c,filter){
	//render table,get table's data,checks,checkresult etc.
	c.RenderNGPage=function(args){
		args=args||{};
		args.tmplName = fmt.Sprintf("%s/%s.html",c.ControllerName,c.ActionName);
		c.RenderTemplate("ngpage.html",args);
	};
	c.RenderProjectPage=function(args){
		args=args||{};
		args.tmplName = fmt.Sprintf("%s/%s.html",c.ControllerName,c.ActionName);
		c.RenderTemplate("pjpage.html",args);
	};
	c.RenderMDPage=function(args){
		args=args||{};
		args.tmplName = fmt.Sprintf("%s/%s.md",c.ControllerName,c.ActionName);
		c.RenderTemplate("mdpage.html",args);
	};
	c.RenderRecordView=function(args){
		args=args||{};
		c.RenderTemplate("recordview.html",args);
	};
  /*c.RenderModelOperate=function(optName,keys,args){
    args=args||{};
    opt = c.Model("rt_modeloperate");
    opt.FillByID(optName);
    if(opt.RowCount() == 0){
      throw fmt.Sprintf("the bill opearte %q not found at datatable [rt_billoperate]",boptName);
    }

    dat = opt.GetRow(0);
    if(!c.GradeCanUse(dat.grade)){
      throw fmt.Sprintf("the bill opearte %q grade is %q,can't use by grade:%q",optName,dat.grade,c.CurrentGrade);
    }
    tabData = c.Model(dat.modelname);
    if(dat.operatetype != pc.TABLE_ADD){
      if(!keys){
        throw fmt.Sprintf("the bill opearte %q is not TABLE_ADD,but no keys passed",optName);
      }
      tabData.FillByID(keys);
      if(tabData.RowCount() == 0){
        throw fmt.Sprintf("the bill %q record not exist,keys:%v",dat.modelname,tableKeys);
      }
      args.Data = tabData.GetRow(0);
    }else{
      args.Data = tabData.NewRow();
    }
    args.Title = optName;
    args.ModelName = dat.modelname;
    args.OperateType = dat.operatetype;
    args.FillValue = dat.fillvalue||{};
    args.Permissions = dat.permissions||{};
    //must clone the function call result,otherwise throw exception
    args.Checks =_.map(c.ModelChecks(dat.modelname),function(chk,idx,list){
      newChk = _.omit(chk,"SqlWhere","Grade");
      if(newChk.RunAtServer){
        delete newChk.Script;
      }
      return newChk
    });
    chkResult = c.Model("rt_checkresult");
    chkResult.FillWhere("tablename=$1 and pks=$2 and $3 like grade||'%'",dat.modelname,keys,c.CurrentGrade);
    args.CheckResult = chkResult.Rows();
    columns ={};
    _.each(tabData.Columns(),function(val,key,list){
      columns[key] = {
        Name:val.Name,
        DataType:val.PGType.Type,
        MaxSize:val.PGType.MaxSize,
        NotNull:val.PGType.NotNull
      };
    });
    args.Fields = columns;
    c.RenderTemplate("table",args);
  };*/

  if(!c.HasResult() && filter.length>0){
    filter[0](c,filter.slice(1));
  }

}
