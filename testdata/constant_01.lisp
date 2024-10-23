(defcolumns X Y)
;; X == Y + n - n
(defconstraint c1 () (- X Y (+ 1 1) (- 0 2)))
(defconstraint c2 () (- X Y (+ 1 1 1) (- 0 1 2)))
(defconstraint c3 () (- X Y (+ 2 1) (- 0 2 1)))
