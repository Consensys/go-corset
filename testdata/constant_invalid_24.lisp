;;error:6:48-51:not permitted in const context
(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (ST :i16))
(defconst (ONE :extern) (^ -2 0))
(defconstraint c1 () (vanishes! (* ST (shift X ONE))))
