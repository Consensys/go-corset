;;error:6:42-45:not permitted in pure context
;;error:6:76-79:not permitted in pure context
;;error:6:22-83:expected bool, found i20
(defcolumns (CT :i4) (X :i16))
(defun (ONE) X)
(defconstraint c1 () (* (- CT (shift CT (ONE))) (- (+ CT (ONE)) (shift CT (ONE)))))
