;;error:5:32-37:not permitted in pure context
;;error:5:22-39:expected loobean constraint (found 𝔽)
(defcolumns X)
(defun (ONE) 1)
(defconstraint c1 () (* X (^ 2 (ONE))))
