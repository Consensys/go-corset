(defconst ONE 1)
(defcolumns CT)
(defconstraint c1 () (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE))))
