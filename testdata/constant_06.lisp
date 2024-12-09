(defconst TWO 2)
(defcolumns X Y)
;; Y == X*X
(defconstraint c1 () (- Y (^ X TWO)))
