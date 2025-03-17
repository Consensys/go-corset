;;error:7:43-46:not permitted in const context
(defpurefun (vanishes! x) (== 0 x))

(defconst (TWO :extern) 2)
(defcolumns (X :i16) (Y :i16))
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
