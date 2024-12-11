(defcolumns X)
(defun ((prevX :@loob)) (shift X -1))
(defconstraint c1 () (prevX))
