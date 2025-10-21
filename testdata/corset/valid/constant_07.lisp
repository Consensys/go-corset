(defcolumns (X :i16) (ST :i16))
(defconst ONE (^ -2 0))
(defconstraint c1 () (== 0 (* ST (shift X ONE))))
