(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i32) (Y :i32))
(defconstraint c1 () (vanishes! (+ X (- 0 Y))))
(defconstraint c2 () (vanishes! (+ (- 0 Y) X)))
