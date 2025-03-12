(defcolumns (X :i16) (Y :i16))
(defun (Xmul x) (* X x))
(defpurefun ((eq :i16@loob) x y) (- x y))
;; Y == 2 * X
(defconstraint c1 () (eq Y (Xmul 2)))
