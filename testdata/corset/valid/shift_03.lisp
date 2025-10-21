(defcolumns (X :i16) (Y :i16) (ST :i16))
(defconstraint c1 () (== 0 (* ST (- (shift X 1) (+ Y Y)))))
