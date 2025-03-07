(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16))
(defun (double x) (+ x x))
(defpurefun (eq x y) (- x y))
(defunalias times2 double)
;; Y == 2 * X
(defconstraint c1 () (vanishes! (eq Y (times2 X))))
