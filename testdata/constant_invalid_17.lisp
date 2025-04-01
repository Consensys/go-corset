;;error:5:32-35:not permitted in pure context
;;error:5:22-37:expected bool, found u16
(defcolumns (X :i16))
(defun (ONE) 1)
(defconstraint c1 () (shift X (ONE)))
