;;error:3:28-37:incorrect number of arguments
(defcolumns (X :i16) (ST :i16))
(defconstraint c1 () (* ST (shift X)))
