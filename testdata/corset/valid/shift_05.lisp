(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (* (shift X 1) Y)))
