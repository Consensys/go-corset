(defpurefun (vanishes! x) (== 0 x))

(defcolumns (_X :i16) (_Y :i16))
(defconstraint c1 () (vanishes! (- _X _Y)))
(defconstraint c2 () (vanishes! (- _Y _X)))
