(module m1)
(defcolumns (X :i16))
(module m2)
(defcolumns (X :i16) (Y :i16))
(defconstraint heartbeat () (== 0 (* X Y)))
