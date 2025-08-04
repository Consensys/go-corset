(module test)
(defcolumns (X :i16))
(defconstraint heartbeat () (== 0 X))
