select
	*
from
	lx_rvterm
where
	access='private' and
	username = {{str .UserName}}
