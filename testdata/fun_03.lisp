(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns (X :i16) (ST :i16))
(defun (get) X)
(defconstraint c1 ()
  (vanishes! (* ST (shift (get) 1))))
