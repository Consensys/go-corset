;;error:3:37-38:not permitted in pure context
(defcolumns (X :i16@loob) (ST :i16@loob))
(defconstraint c1 () (* ST (shift X X)))
