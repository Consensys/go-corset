(defcolumns X Y)
(defun (Xm1) (- X 1))
(defpurefun ((eq :@loob) x y) (- x y))
;; Y == X
(defconstraint c1 () (eq (- Y 1) (Xm1)))
