(defcolumns (X :i16) (Y :i16))
(defun (double x) (+ x x))
(defpurefun ((eq :bool) x y) (== x y))

;; Y == 2 * X
(defconstraint c1 () (eq Y (double X)))
