(defcolumns (X :i16) (Y :i16))
(defun (Xm1) (- X 1))
(defpurefun ((eq :i16) x y) (- x y))
;; Y == X
(defconstraint c1 () (eq (- Y 1) (Xm1)))
