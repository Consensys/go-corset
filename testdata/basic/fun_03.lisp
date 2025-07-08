(defpurefun (vanishes! x) (== 0 x))
;;
(defcolumns (X :i16) (ST :i16))
(defun (get) X)
(defconstraint c1 ()
  (vanishes! (* ST (shift (get) 1))))
