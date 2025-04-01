(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i16) (A :i16))
(defconstraint c1 () (vanishes! (* ST (- 1 (~ A)))))
