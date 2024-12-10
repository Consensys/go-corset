(defcolumns X)
(defun (prevX) (shift X -1))
(defconstraint c1 () (prevX))
