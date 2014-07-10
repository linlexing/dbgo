do $DO$
declare
  src text;
  cursDBPatch CURSOR FOR SELECT script FROM <#.TempTableName#>;
begin
  for one in cursDBPatch loop
    execute one.script;
  end loop;
end;
$DO$
