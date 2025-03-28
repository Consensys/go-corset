(defpurefun (vanishes! x) (== 0 x))

(defconst TWO 2)
(defcolumns (X :i16) (Y :i32))
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
