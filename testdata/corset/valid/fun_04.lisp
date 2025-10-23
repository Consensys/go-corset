;;
(defcolumns (X :i16))
(defun (get) X)
(defconstraint c1 ()
  (== 0 (shift (get) -1)))
