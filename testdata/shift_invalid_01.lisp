;;error:3:37-38:not permitted in pure context
(defcolumns (X :i16) (ST :i16))
(defconstraint c1 () (* ST (shift X X)))
