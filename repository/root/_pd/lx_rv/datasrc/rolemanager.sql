select *
from lx_role a
where grade_canuse({{str .CurrentGrade}},a.grade)
