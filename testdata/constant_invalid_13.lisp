(defcolumns CT ONE)
(defconstraint c1 () (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE))))
