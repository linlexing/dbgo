{{if eq .DriverName "postgres"}}
DROP FUNCTION IF EXISTS grade_canuse(text, text);
go
CREATE FUNCTION grade_canuse(current_grade text, canuse_grade text)
  RETURNS boolean AS
$BODY$
BEGIN
	IF current_grade = '' THEN
	  RETURN true;
	END IF;
	IF current_grade like canuse_grade||'%' THEN
	  RETURN true;
	END IF;
	RETURN false;
END;
$BODY$
  LANGUAGE plpgsql IMMUTABLE;
{{else if eq .DriverName "mysql"}}
DROP FUNCTION IF EXISTS grade_canuse
go
CREATE FUNCTION grade_canuse(current_grade text, canuse_grade text) RETURNS bit(1)
    DETERMINISTIC
BEGIN
	IF current_grade = '' THEN
	  RETURN true;
	END IF;
	IF current_grade like concat(canuse_grade,'%') THEN
	  RETURN true;
	END IF;
	RETURN false;
END
{{end}}
