;;error:5:33-36:not permitted in pure context
;;error:5:22-39:expected bool, found int
(defcolumns (X :i16))
(defun (ONE) 1)
(defconstraint c1 () (* X (^ 2 (ONE))))
