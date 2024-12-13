(defpurefun ((vanishes! :@loob) x) x)
;;
(defcolumns X)
(defun (get) X)
(defconstraint c1 ()
  (vanishes! (shift (get) -1)))
