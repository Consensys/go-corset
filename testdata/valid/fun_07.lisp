(defcolumns (X :i16))
(defun (getX) X)
(defcomputedcolumn (Y :i16) (getX))
