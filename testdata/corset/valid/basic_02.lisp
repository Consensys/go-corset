(defcolumns (X :i32) (Y :i32))
(defconstraint c1 () (== 0 (+ X (- 0 Y))))
(defconstraint c2 () (== 0 (+ (- 0 Y) X)))
