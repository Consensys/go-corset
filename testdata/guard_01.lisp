(defcolumns STAMP X)
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard STAMP) (* (- X 1) (- X 2)))
