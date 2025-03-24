;;error:6:33-36:not permitted in pure context
;;error:6:22-39:expected bool, found int
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (X :i16) (Y :i16) (Z :i16))
(defun (TWO) Z)
(defconstraint c1 () (- Y (^ X (TWO))))
