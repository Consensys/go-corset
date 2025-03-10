;;error:3:26-27:invalid condition (neither loobean nor boolean)
(defcolumns (A :i16) (B :i16) (C :i16))
(defconstraint c1 () (if A B C))
