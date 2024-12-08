(defcolumns CT X)
(defun (ONE) X)
(defconstraint c1 () (* (- CT (shift CT (ONE))) (- (+ CT (ONE)) (shift CT (ONE)))))
