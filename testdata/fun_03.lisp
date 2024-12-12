(defpurefun ((vanishes! :@loob) x) x)
;;
(defcolumns X ST)
(defun (get) X)
(defconstraint c1 ()
  (vanishes! (* ST (shift (get) 1))))
