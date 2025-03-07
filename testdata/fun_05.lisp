(defcolumns X)
(defun ((prevX :i16@loob)) (shift X -1))
(defconstraint c1 () (prevX))
