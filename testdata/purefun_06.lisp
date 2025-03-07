(defcolumns X Y)
(defun (double@d x) (+ x x))
(defpurefun ((~eq :i16@loob) x y) (~ (- x y)))

;; Y == 2 * X
(defconstraint c1 () (~eq Y (double@d X)))
