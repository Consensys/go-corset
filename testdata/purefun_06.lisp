(defcolumns X Y)
(defun (double x) (+ x x))
(defpurefun ((~eq :@loob) x y) (~ (- x y)))

;; Y == 2 * X
(defconstraint c1 () (~eq Y (double X)))
