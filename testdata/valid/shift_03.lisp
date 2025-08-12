(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (ST :i16))
(defconstraint c1 () (vanishes! (* ST (- (shift X 1) (+ Y Y)))))
