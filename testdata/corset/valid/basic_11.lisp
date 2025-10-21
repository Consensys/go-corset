(defcolumns (_X :i16) (_Y :i16))
(defconstraint c1 () (== 0 (- _X _Y)))
(defconstraint c2 () (== 0 (- _Y _X)))
