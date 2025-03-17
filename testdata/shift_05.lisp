(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (* (shift X 1) Y)))
