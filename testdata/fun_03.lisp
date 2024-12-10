(defcolumns X ST)
(defun (get) X)
(defconstraint c1 () (* ST (shift (get) 1)))
