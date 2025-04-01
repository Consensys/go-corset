(defpurefun (vanishes! x) (== 0 x))

(defcolumns (A :i16))
(defconstraint c1 () (vanishes! (- 2 (~ A))))
