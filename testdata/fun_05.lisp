(defcolumns (X :i16))
(defun ((prevX :i16)) (shift X -1))
(defconstraint c1 () (prevX))
