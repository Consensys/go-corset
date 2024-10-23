(defcolumns X Y)
;; Y == X*X
(defconstraint c1 () (- Y (^ X 2)))
