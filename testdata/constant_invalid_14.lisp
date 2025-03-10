;;error:6:41-46:not permitted in pure context
;;error:6:75-80:not permitted in pure context
;;error:6:22-83:expected loobean constraint (found u16)
(defcolumns (CT :i4) (X :i16))
(defun (ONE) X)
(defconstraint c1 () (* (- CT (shift CT (ONE))) (- (+ CT (ONE)) (shift CT (ONE)))))
