(defcolumns (X :i256) (Y :i256))
;; X == Y + 1
(defconstraint c1 () (if (!= 0 X) (== X (+ Y 1))))
