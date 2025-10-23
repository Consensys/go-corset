;;
(defcolumns (X :i16) (ST :i16))
(defun (get) X)
(defconstraint c1 ()
  (== 0 (* ST (shift (get) 1))))
