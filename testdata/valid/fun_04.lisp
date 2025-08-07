(defpurefun (vanishes! x) (== 0 x))
;;
(defcolumns (X :i16))
(defun (get) X)
(defconstraint c1 ()
  (vanishes! (shift (get) -1)))
