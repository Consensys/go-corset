(defcolumns CT)
(defconstraint c1 () (* (- CT (shift CT 1)) (- (+ CT 1) (shift CT 1))))
