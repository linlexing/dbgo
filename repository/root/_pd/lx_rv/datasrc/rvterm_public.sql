select
	*
from
	lx_rvterm
where
	access='public' and
	grade_canuse({{str .CurrentGrade}},grade)
