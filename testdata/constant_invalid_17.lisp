;;error:5:31-36:not permitted in pure context
;;error:5:22-37:expected loobean constraint (found 𝔽)
(defcolumns X)
(defun (ONE) 1)
(defconstraint c1 () (shift X (ONE)))
