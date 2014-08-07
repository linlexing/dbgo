{{if eq .DriverName "postgres"}}
DROP FUNCTION IF EXISTS strcat(text, text);
go
CREATE FUNCTION strcat(str1 text, str2 text)
  RETURNS text AS
$BODY$
BEGIN
	RETURN concat(str1,str2);
END;
$BODY$
  LANGUAGE plpgsql IMMUTABLE;
{{else if eq .DriverName "mysql"}}
DROP FUNCTION IF EXISTS strcat
go
CREATE FUNCTION strcat(str1 text, str2 text) RETURNS text
    DETERMINISTIC
BEGIN
	RETURN concat(str1,str2);
END
{{end}}
