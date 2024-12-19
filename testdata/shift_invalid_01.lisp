;;error:3:37-38:not permitted in pure context
(defcolumns (X :@loob) (ST :@loob))
(defconstraint c1 () (* ST (shift X X)))
