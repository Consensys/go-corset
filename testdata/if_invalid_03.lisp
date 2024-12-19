;;error:3:26-27:invalid condition (neither loobean nor boolean)
(defcolumns A B C)
(defconstraint c1 () (if A B C))
