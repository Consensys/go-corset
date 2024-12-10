(defcolumns X Y)
(defun (Xmul x) (* X x))
(defpurefun (eq x y) (- x y))
;; Y == 2 * X
(defconstraint c1 () (eq Y (Xmul 2)))
